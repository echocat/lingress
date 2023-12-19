package lingress

import (
	"crypto/tls"
	"fmt"
	lctx "github.com/echocat/lingress/context"
	"github.com/echocat/lingress/fallback"
	"github.com/echocat/lingress/file/providers"
	"github.com/echocat/lingress/management"
	"github.com/echocat/lingress/proxy"
	"github.com/echocat/lingress/rules"
	"github.com/echocat/lingress/server"
	"github.com/echocat/lingress/settings"
	"github.com/echocat/lingress/support"
	"github.com/echocat/slf4g"
	"github.com/echocat/slf4g/level"
	"net"
	"net/http"
	"reflect"
	"strings"
	"sync/atomic"
)

type Lingress struct {
	settings *settings.Settings

	RulesRepository rules.CombinedRepository
	Proxy           *proxy.Proxy
	Fallback        *fallback.Fallback
	Management      *management.Management
	Http            *server.HttpConnector
	Https           *server.HttpConnector

	accessLogQueue chan *accessLogEntry

	unprocessableConnectionDocumented map[reflect.Type]bool

	logger                 log.Logger
	accessLogger           log.Logger
	accessLoggerMessageKey string
}

type ConnectionStates struct {
	New    uint64
	Active uint64
	Idle   uint64
}

func New(s *settings.Settings, fps providers.FileProviders) (*Lingress, error) {
	logProvider := log.GetProvider()

	connectorIds := []server.ConnectorId{server.DefaultConnectorIdHttp, server.DefaultConnectorIdHttps}
	r, err := rules.NewRepository(s, logProvider.GetLogger("rules"))
	if err != nil {
		return nil, err
	}
	p, err := proxy.New(s, r, logProvider.GetLogger("proxy"))
	if err != nil {
		return nil, err
	}
	f, err := fallback.New(s, fps, logProvider.GetLogger("fallback"))
	if err != nil {
		return nil, err
	}
	m, err := management.New(s, connectorIds, r, logProvider.GetLogger("management"))
	if err != nil {
		return nil, err
	}
	hHttp, err := server.NewHttpConnector(s, server.DefaultConnectorIdHttp, logProvider.GetLogger(string(server.DefaultConnectorIdHttp)))
	if err != nil {
		return nil, err
	}
	hHttps, err := server.NewHttpConnector(s, server.DefaultConnectorIdHttps, logProvider.GetLogger(string(server.DefaultConnectorIdHttps)))
	if err != nil {
		return nil, err
	}

	result := &Lingress{
		settings: s,

		RulesRepository: r,
		Proxy:           p,
		Fallback:        f,
		Management:      m,
		Http:            hHttp,
		Https:           hHttps,

		accessLogger: logProvider.GetLogger("accessLog"),
		logger:       logProvider.GetLogger("core"),
	}

	result.accessLoggerMessageKey = result.accessLogger.GetProvider().GetFieldKeysSpec().GetMessage()
	result.Http.Handler = result
	result.Https.Handler = result

	p.ResultHandler = result.onResult
	p.AccessLogger = result.onAccessLog
	p.MetricsCollector = result.Management

	return result, nil
}

func (this *Lingress) ServeHTTP(connector server.Connector, resp http.ResponseWriter, req *http.Request) {
	finalize := this.Management.CollectClientStarted(connector.GetId())
	defer finalize()

	this.Proxy.ServeHTTP(connector, resp, req)
}

func (this *Lingress) OnConnState(connector server.Connector, conn net.Conn, state http.ConnState) {
	annotated, ok := conn.RemoteAddr().(server.AnnotatedAddr)
	if !ok {
		connType := reflect.TypeOf(conn)
		if !this.unprocessableConnectionDocumented[connType] {
			this.unprocessableConnectionDocumented[connType] = true
			this.logger.
				With("connType", connType.String()).
				Error("Cannot inspect connection of provided connection type; for those kinds of connections there will be no statistics be available.")
		}
		return
	}

	previous := annotated.GetState()
	annotated.SetState(&state)

	if previous != nil && *previous == state {
		return
	}

	client, ok := this.Management.Metrics.Client[connector.GetId()]
	if !ok {
		return
	}

	source := client.Connections.Source
	if previous != nil {
		switch *previous {
		case http.StateNew:
			atomic.AddUint64(&source.New, ^uint64(0))
		case http.StateActive:
			atomic.AddUint64(&source.Active, ^uint64(0))
		case http.StateIdle:
			atomic.AddUint64(&source.Idle, ^uint64(0))
		default:
			return
		}
	}

	switch state {
	case http.StateNew:
		atomic.AddUint64(&source.New, 1)
		atomic.AddUint64(&source.Current, 1)
		atomic.AddUint64(&source.Total, 1)
	case http.StateActive:
		atomic.AddUint64(&source.Active, 1)
	case http.StateIdle:
		atomic.AddUint64(&source.Idle, 1)
	default:
		atomic.AddUint64(&source.Current, ^uint64(0))
	}

	return
}

func (this *Lingress) Init(stop support.Channel) error {
	if err := this.RulesRepository.Init(stop); err != nil {
		return err
	}
	if err := this.Proxy.Init(stop); err != nil {
		return err
	}

	if s := this.settings.AccessLog.QueueSize; s > 0 {
		queue := make(chan *accessLogEntry, this.settings.AccessLog.QueueSize)
		go this.accessLogQueueWorker(stop, queue)
		this.accessLogQueue = queue
	} else {
		this.accessLogQueue = nil
	}

	this.unprocessableConnectionDocumented = make(map[reflect.Type]bool)

	go this.shutdownListener(stop)

	if tlsConfig, err := this.createTlsConfig(); err != nil {
		return err
	} else {
		this.Https.Server.TLSConfig = tlsConfig
	}

	if err := this.Http.Serve(stop); err != nil {
		return err
	}
	if err := this.Https.Serve(stop); err != nil {
		return err
	}

	if err := this.Management.Init(stop); err != nil {
		return err
	}

	return nil
}

func (this *Lingress) createTlsConfig() (*tls.Config, error) {
	fail := func(err error) (*tls.Config, error) {
		return nil, fmt.Errorf("cannot create TLS config: %w", err)
	}

	result := tls.Config{
		Certificates:   []tls.Certificate{},
		GetCertificate: this.resolveCertificate,
	}

	if this.settings.Tls.FallbackCertificate.Get() {
		v, err := support.CreateDummyCertificate()
		if err != nil {
			return fail(err)
		}
		result.Certificates = append(result.Certificates, v)
	}

	return &result, nil
}

func (this *Lingress) resolveCertificate(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
	var query rules.CertificateQuery
	if err := query.Host.Set(info.ServerName); err != nil {
		return nil, nil
	}

	certificates, err := this.RulesRepository.FindCertificatesBy(query)
	if err != nil {
		return nil, err
	}

	for _, certificate := range certificates {
		if err := info.SupportsCertificate(certificate); err == nil {
			return certificate, nil
		}
	}

	return nil, nil
}

func (this *Lingress) shutdownListener(stop support.Channel) {
	stop.Wait()
	this.Http.Shutdown()
	this.Https.Shutdown()
}

func (this *Lingress) accessLogQueueWorker(stop support.Channel, queue chan *accessLogEntry) {
	stopCh := support.ToChan(stop)
	for {
		select {
		case entry := <-queue:
			this.doAccessLog(entry)
		case <-stopCh:
			return
		}
	}
}

func (this *Lingress) onAccessLog(ctx *lctx.Context) {
	if log.IsLevelEnabled(this.logLevelByContext(ctx)) {
		entry := &accessLogEntry{
			Context: ctx,
		}
		if q := this.accessLogQueue; q != nil {
			q <- entry
		} else {
			this.doAccessLog(entry)
		}
	}
}

func (this *Lingress) doAccessLog(entry *accessLogEntry) {
	defer func() {
		if err := entry.Release(); err != nil {
			this.logger.
				WithError(err).
				Error("Problem while releasing context.")
		}
	}()

	entry.load(this.settings.AccessLog.Inline.Get())
	status := 0
	if c, ok := entry.data["client"]; ok {
		if ps, ok := c.(map[string]interface{})["status"]; ok {
			if is, ok := ps.(int); ok {
				status = is
			}
		}
	}
	if !this.shouldBeLogged(entry) && status < 400 {
		return
	}

	lvl := this.logLevelByStatus(status)
	event := this.accessLogger.NewEvent(lvl, entry.data)
	this.accessLogger.Log(event, 0)
}

func (this *Lingress) shouldBeLogged(entry *accessLogEntry) bool {
	userAgent, remotes := this.userAgentAndRemoteOf(entry)
	if strings.HasPrefix(userAgent, "kube-probe/") && this.hasPrivateNetworkIp(remotes) {
		return false
	}
	return true
}

func (this *Lingress) userAgentAndRemoteOf(entry *accessLogEntry) (userAgent string, remotes []net.IP) {
	c, ok := entry.data["client"]
	if !ok {
		return
	}
	if pua, ok := c.(map[string]interface{})["userAgent"]; !ok {
		userAgent = ""
	} else if ua, ok := pua.(string); !ok {
		userAgent = ""
	} else {
		userAgent = ua
	}

	if pr, ok := c.(map[string]interface{})["address"]; !ok {
		remotes = []net.IP{}
	} else if r, ok := pr.(string); !ok {
		remotes = []net.IP{}
	} else if ips, err := net.LookupIP(r); err != nil {
		remotes = []net.IP{}
	} else {
		remotes = ips
	}

	return
}

func (this *Lingress) hasPrivateNetworkIp(ips []net.IP) bool {
	for _, ip := range ips {
		for _, pn := range privateNetworks {
			if pn.Contains(ip) {
				return true
			}
		}
	}
	return false
}

func (this *Lingress) logLevelByContext(ctx *lctx.Context) level.Level {
	if err := ctx.Error; err != nil && ctx.Client.Status <= 0 {
		ctx.Client.Status = 500
	}
	return this.logLevelByStatus(ctx.Client.Status)
}

func (this *Lingress) logLevelByStatus(status int) level.Level {
	if status < 500 {
		return level.Info
	} else if status == http.StatusBadGateway ||
		status == http.StatusServiceUnavailable ||
		status == http.StatusGatewayTimeout {
		return level.Warn
	} else {
		return level.Error
	}
}

func (this *Lingress) onResult(ctx *lctx.Context) {
	if ctx.Result.WasResponseSendToClient() {
		return
	}

	ctx.Stage = lctx.StagePrepareClientResponse
	if proceed, err := this.Proxy.Interceptors.Handle(ctx); err != nil {
		ctx.Log().WithError(err).Error("Cannot prepare error response.")
		return
	} else if !proceed {
		return
	}

	ctx.Stage = lctx.StageSendResponseToClient
	ctx.Client.Status = ctx.Result.Status()
	if ctx.Result == lctx.ResultFailedWithRuleNotFound {
		this.Fallback.Unknown(ctx)
	} else if rr, ok := ctx.Result.(lctx.RedirectResult); ok {
		this.Fallback.Redirect(ctx, ctx.Client.Status, rr.Target)
	} else {
		p := ""
		if u, err := ctx.Client.RequestedUrl(); err == nil && u != nil {
			p = u.Path
		}
		canHandleTemporary := ctx.Client.Request.Method == "GET"
		this.Fallback.Status(ctx, ctx.Client.Status, p, canHandleTemporary)
	}

	ctx.Stage = lctx.StageDone
}

type accessLogEntry struct {
	*lctx.Context
	data map[string]interface{}
}

func (this *accessLogEntry) load(inline bool) {
	this.data = this.AsMap(inline)
}

func (this *accessLogEntry) getField(sub string, inlined bool, field string) interface{} {
	if sub == "" {
		return this.data[field]
	}

	if inlined {
		return this.data[sub+"."+field]
	}

	if sm, ok := this.data[sub].(map[string]interface{}); ok {
		return sm[field]
	}

	return nil
}
