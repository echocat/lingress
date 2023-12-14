package lingress

import (
	"crypto/tls"
	"fmt"
	lctx "github.com/echocat/lingress/context"
	"github.com/echocat/lingress/fallback"
	"github.com/echocat/lingress/management"
	"github.com/echocat/lingress/proxy"
	"github.com/echocat/lingress/rules"
	"github.com/echocat/lingress/server"
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
	RulesRepository    rules.CombinedRepository
	Proxy              *proxy.Proxy
	Fallback           *fallback.Fallback
	Management         *management.Management
	Http               *server.HttpConnector
	Https              *server.HttpConnector
	AccessLogQueueSize uint16
	InlineAccessLog    bool

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

func New(fps support.FileProviders) (*Lingress, error) {
	logProvider := log.GetProvider()

	connectorIds := []server.ConnectorId{server.DefaultConnectorIdHttp, server.DefaultConnectorIdHttps}
	if r, err := rules.NewRepository(logProvider.GetLogger("rules")); err != nil {
		return nil, err
	} else if p, err := proxy.New(r, logProvider.GetLogger("proxy")); err != nil {
		return nil, err
	} else if f, err := fallback.New(fps, logProvider.GetLogger("fallback")); err != nil {
		return nil, err
	} else if m, err := management.New(connectorIds, r, logProvider.GetLogger("management")); err != nil {
		return nil, err
	} else if hHttp, err := server.NewHttpConnector(server.DefaultConnectorIdHttp, logProvider.GetLogger("http")); err != nil {
		return nil, err
	} else if hHttps, err := server.NewHttpConnector(server.DefaultConnectorIdHttps, logProvider.GetLogger("https")); err != nil {
		return nil, err
	} else {
		result := &Lingress{
			RulesRepository:    r,
			Proxy:              p,
			Fallback:           f,
			Management:         m,
			Http:               hHttp,
			Https:              hHttps,
			AccessLogQueueSize: 5000,

			accessLogger: logProvider.GetLogger("accessLog"),
			logger:       logProvider.GetLogger("core"),
		}

		result.accessLoggerMessageKey = result.accessLogger.GetProvider().GetFieldKeysSpec().GetMessage()
		result.Http.Handler = result
		result.Http.MaxConnections = 256

		result.Https.Handler = result
		result.Https.Server.Addr = ":8443"

		p.ResultHandler = result.onResult
		p.AccessLogger = result.onAccessLog
		p.MetricsCollector = result.Management

		return result, nil
	}
}

func (instance *Lingress) RegisterFlag(fe support.FlagEnabled, appPrefix string) error {
	fe.Flag("accessLog.queueSize", "Maximum number of accessLog elements that could be queue before blocking.").
		PlaceHolder(fmt.Sprint(instance.AccessLogQueueSize)).
		Envar(support.FlagEnvName(appPrefix, "ACCESS_LOG_QUEUE_SIZE")).
		Uint16Var(&instance.AccessLogQueueSize)
	fe.Flag("accessLog.inline", "Instead of exploding the accessLog entries into sub-entries everything is inlined into the root object.").
		Envar(support.FlagEnvName(appPrefix, "ACCESS_LOG_INLINE")).
		BoolVar(&instance.InlineAccessLog)
	if err := instance.RulesRepository.RegisterFlag(fe, appPrefix); err != nil {
		return err
	}
	if err := instance.Proxy.RegisterFlag(fe, appPrefix); err != nil {
		return err
	}
	if err := instance.Fallback.RegisterFlag(fe, appPrefix); err != nil {
		return err
	}
	if err := instance.Management.RegisterFlag(fe, appPrefix); err != nil {
		return err
	}
	if err := instance.Http.RegisterFlag(fe, appPrefix); err != nil {
		return err
	}
	if err := instance.Https.RegisterFlag(fe, appPrefix); err != nil {
		return err
	}
	return nil
}

func (instance *Lingress) ServeHTTP(connector server.Connector, resp http.ResponseWriter, req *http.Request) {
	finalize := instance.Management.CollectClientStarted(connector.GetId())
	defer finalize()

	instance.Proxy.ServeHTTP(connector, resp, req)
}

func (instance *Lingress) OnConnState(connector server.Connector, conn net.Conn, state http.ConnState) {
	annotated, ok := conn.RemoteAddr().(server.AnnotatedAddr)
	if !ok {
		connType := reflect.TypeOf(conn)
		if !instance.unprocessableConnectionDocumented[connType] {
			instance.unprocessableConnectionDocumented[connType] = true
			instance.logger.
				With("connType", connType.String()).
				Error("cannot inspect connection of provided connection type; for those kinds of connections there will be no statistics be available")
		}
		return
	}

	previous := annotated.GetState()
	annotated.SetState(&state)

	if previous != nil && *previous == state {
		return
	}

	client, ok := instance.Management.Metrics.Client[connector.GetId()]
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

func (instance *Lingress) Init(stop support.Channel) error {
	if err := instance.RulesRepository.Init(stop); err != nil {
		return err
	}
	if err := instance.Proxy.Init(stop); err != nil {
		return err
	}

	if s := instance.AccessLogQueueSize; s > 0 {
		queue := make(chan *accessLogEntry, instance.AccessLogQueueSize)
		go instance.accessLogQueueWorker(stop, queue)
		instance.accessLogQueue = queue
	} else {
		instance.accessLogQueue = nil
	}

	instance.unprocessableConnectionDocumented = make(map[reflect.Type]bool)

	go instance.shutdownListener(stop)

	if tlsConfig, err := instance.createTlsConfig(); err != nil {
		return err
	} else {
		instance.Https.Server.TLSConfig = tlsConfig
	}

	if err := instance.Http.Serve(stop); err != nil {
		return err
	}
	if err := instance.Https.Serve(stop); err != nil {
		return err
	}

	if err := instance.Management.Init(stop); err != nil {
		return err
	}

	return nil
}

func (instance *Lingress) createTlsConfig() (*tls.Config, error) {
	fail := func(err error) (*tls.Config, error) {
		return nil, fmt.Errorf("cannot create TLS config: %w", err)
	}
	defaultCert, err := support.CreateDummyCertificate()
	if err != nil {
		return fail(err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{
			defaultCert,
		},
		GetCertificate: instance.resolveCertificate,
	}, nil
}

func (instance *Lingress) resolveCertificate(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
	certificates, err := instance.RulesRepository.FindCertificatesBy(rules.CertificateQuery{
		Host: info.ServerName,
	})
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

func (instance *Lingress) shutdownListener(stop support.Channel) {
	stop.Wait()
	instance.Http.Shutdown()
	instance.Https.Shutdown()
}

func (instance *Lingress) accessLogQueueWorker(stop support.Channel, queue chan *accessLogEntry) {
	stopCh := support.ToChan(stop)
	for {
		select {
		case entry := <-queue:
			instance.doAccessLog(entry)
		case <-stopCh:
			return
		}
	}
}

func (instance *Lingress) onAccessLog(ctx *lctx.Context) {
	if log.IsLevelEnabled(instance.logLevelByContext(ctx)) {
		entry := &accessLogEntry{
			Context: ctx,
		}
		if q := instance.accessLogQueue; q != nil {
			q <- entry
		} else {
			instance.doAccessLog(entry)
		}
	}
}

func (instance *Lingress) doAccessLog(entry *accessLogEntry) {
	defer entry.Release()

	entry.load(instance.InlineAccessLog)
	status := 0
	if c, ok := entry.data["client"]; ok {
		if ps, ok := c.(map[string]interface{})["status"]; ok {
			if is, ok := ps.(int); ok {
				status = is
			}
		}
	}
	if !instance.shouldBeLogged(entry) && status < 400 {
		return
	}

	lvl := instance.logLevelByStatus(status)
	event := instance.accessLogger.NewEvent(lvl, entry.data)
	instance.accessLogger.Log(event, 0)
}

func (instance *Lingress) shouldBeLogged(entry *accessLogEntry) bool {
	userAgent, remotes := instance.userAgentAndRemoteOf(entry)
	if strings.HasPrefix(userAgent, "kube-probe/") && instance.hasPrivateNetworkIp(remotes) {
		return false
	}
	return true
}

func (instance *Lingress) userAgentAndRemoteOf(entry *accessLogEntry) (userAgent string, remotes []net.IP) {
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

func (instance *Lingress) hasPrivateNetworkIp(ips []net.IP) bool {
	for _, ip := range ips {
		for _, pn := range privateNetworks {
			if pn.Contains(ip) {
				return true
			}
		}
	}
	return false
}

func (instance *Lingress) logLevelByContext(ctx *lctx.Context) level.Level {
	if err := ctx.Error; err != nil && ctx.Client.Status <= 0 {
		ctx.Client.Status = 500
	}
	return instance.logLevelByStatus(ctx.Client.Status)
}

func (instance *Lingress) logLevelByStatus(status int) level.Level {
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

func (instance *Lingress) onResult(ctx *lctx.Context) {
	if ctx.Result.WasResponseSendToClient() {
		return
	}

	ctx.Stage = lctx.StagePrepareClientResponse
	if proceed, err := instance.Proxy.Interceptors.Handle(ctx); err != nil {
		ctx.Log().WithError(err).Error("cannot prepare error response")
		return
	} else if !proceed {
		return
	}

	ctx.Stage = lctx.StageSendResponseToClient
	ctx.Client.Status = ctx.Result.Status()
	if ctx.Result == lctx.ResultFailedWithRuleNotFound {
		instance.Fallback.Unknown(ctx)
	} else if rr, ok := ctx.Result.(lctx.RedirectResult); ok {
		instance.Fallback.Redirect(ctx, ctx.Client.Status, rr.Target)
	} else {
		p := ""
		if u, err := ctx.Client.RequestedUrl(); err == nil && u != nil {
			p = u.Path
		}
		canHandleTemporary := ctx.Client.Request.Method == "GET"
		instance.Fallback.Status(ctx, ctx.Client.Status, p, canHandleTemporary)
	}

	ctx.Stage = lctx.StageDone
}

type accessLogEntry struct {
	*lctx.Context
	data map[string]interface{}
}

func (instance *accessLogEntry) load(inline bool) {
	instance.data = instance.AsMap(inline)
}

func (instance *accessLogEntry) getField(sub string, inlined bool, field string) interface{} {
	if sub == "" {
		return instance.data[field]
	}

	if inlined {
		return instance.data[sub+"."+field]
	}

	if sm, ok := instance.data[sub].(map[string]interface{}); ok {
		return sm[field]
	}

	return nil
}
