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
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
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

	accessLogQueue        chan accessLogEntry
	connectionInformation *ConnectionInformation
}

type accessLogEntry map[string]interface{}

type ConnectionStates struct {
	New    uint64
	Active uint64
	Idle   uint64
}

func New(fps support.FileProviders) (*Lingress, error) {
	connectorIds := []server.ConnectorId{server.DefaultConnectorIdHttp, server.DefaultConnectorIdHttps}
	if r, err := rules.NewRepository(); err != nil {
		return nil, err
	} else if p, err := proxy.New(r); err != nil {
		return nil, err
	} else if f, err := fallback.New(fps); err != nil {
		return nil, err
	} else if m, err := management.New(connectorIds, r); err != nil {
		return nil, err
	} else if hHttp, err := server.NewHttpConnector(server.DefaultConnectorIdHttp); err != nil {
		return nil, err
	} else if hHttps, err := server.NewHttpConnector(server.DefaultConnectorIdHttps); err != nil {
		return nil, err
	} else {
		result := &Lingress{
			RulesRepository:       r,
			Proxy:                 p,
			Fallback:              f,
			Management:            m,
			Http:                  hHttp,
			Https:                 hHttps,
			connectionInformation: NewConnectionInformation(),
			AccessLogQueueSize:    5000,
		}

		result.Http.Handler = result
		result.Http.MaxConnections = 512

		result.Https.Handler = result
		result.Https.Server.Addr = ":8443"

		p.ResultHandler = result.onResult
		p.AccessLogger = result.onAccessLog
		p.MetricsCollector = result.Management

		return result, nil
	}
}

func (instance *Lingress) RegisterFlag(fe support.FlagEnabled, appPrefix string) error {
	fe.Flag("accessLogQueueSize", "Maximum number of accessLog elements that could be queue before blocking.").
		PlaceHolder(fmt.Sprint(instance.AccessLogQueueSize)).
		Envar(support.FlagEnvName(appPrefix, "ACCESS_LOG_QUEUE_SIZE")).
		Uint16Var(&instance.AccessLogQueueSize)
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
	previous := instance.connectionInformation.SetState(conn, state, func(target http.ConnState) bool {
		return target != http.StateNew &&
			target != http.StateActive &&
			target != http.StateIdle
	})

	if previous == state {
		return
	}

	client, ok := instance.Management.Metrics.Client[connector.GetId()]
	if !ok {
		return
	}

	source := client.Connections.Source
	switch previous {
	case http.StateNew:
		atomic.AddUint64(&source.New, ^uint64(0))
	case http.StateActive:
		atomic.AddUint64(&source.Active, ^uint64(0))
	case http.StateIdle:
		atomic.AddUint64(&source.Idle, ^uint64(0))
	case -1:
		// ignore
	default:
		return
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
		queue := make(chan accessLogEntry, instance.AccessLogQueueSize)
		go instance.accessLogQueueWorker(stop, queue)
		instance.accessLogQueue = queue
	} else {
		instance.accessLogQueue = nil
	}

	go instance.shutdownListener(stop)

	if tlsConfig, err := instance.createTlsConfig(); err != nil {
		return err
	} else {
		instance.Https.TlsConfig = tlsConfig
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
		return nil, errors.Wrap(err, "cannot create TLS config")
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

func (instance *Lingress) accessLogQueueWorker(stop support.Channel, queue chan accessLogEntry) {
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
		entry := ctx.AsMap()
		if q := instance.accessLogQueue; q != nil {
			q <- entry
		} else {
			instance.doAccessLog(entry)
		}
	}
}

func (instance *Lingress) doAccessLog(entry accessLogEntry) {
	status := 0
	if c, ok := entry["client"]; ok {
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
	log.WithFields(log.Fields(entry)).Log(lvl, "accessLog")
}

func (instance *Lingress) shouldBeLogged(entry accessLogEntry) bool {
	userAgent, remotes := instance.userAgentAndRemoteOf(entry)
	if strings.HasPrefix(userAgent, "kube-probe/") && instance.hasPrivateNetworkIp(remotes) {
		return false
	}
	return true
}

func (instance *Lingress) userAgentAndRemoteOf(entry accessLogEntry) (userAgent string, remotes []net.IP) {
	c, ok := entry["client"]
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

func (instance *Lingress) logLevelByContext(ctx *lctx.Context) log.Level {
	if err := ctx.Error; err != nil && ctx.Client.Status <= 0 {
		ctx.Client.Status = 500
	}
	return instance.logLevelByStatus(ctx.Client.Status)
}

func (instance *Lingress) logLevelByStatus(status int) log.Level {
	if status < 500 {
		return log.InfoLevel
	} else if status == http.StatusBadGateway ||
		status == http.StatusServiceUnavailable ||
		status == http.StatusGatewayTimeout {
		return log.WarnLevel
	} else {
		return log.ErrorLevel
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
