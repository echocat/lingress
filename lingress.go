package lingress

import (
	"context"
	"crypto/tls"
	"fmt"
	lctx "github.com/echocat/lingress/context"
	"github.com/echocat/lingress/fallback"
	"github.com/echocat/lingress/management"
	"github.com/echocat/lingress/proxy"
	"github.com/echocat/lingress/rules"
	"github.com/echocat/lingress/support"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"strings"
	"time"
)

type Lingress struct {
	RulesRepository rules.CombinedRepository
	Proxy           *proxy.Proxy
	Fallback        *fallback.Fallback
	Management      *management.Management
	http            http.Server
	https           http.Server
	accessLogQueue  chan accessLogEntry

	HttpListenAddr    string
	HttpsListenAddr   string
	MaxHeaderBytes    uint
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration

	AccessLogQueueSize uint16
}

type accessLogEntry map[string]interface{}

func New(fps support.FileProviders) (*Lingress, error) {
	if r, err := rules.NewRepository(); err != nil {
		return nil, err
	} else if p, err := proxy.New(r); err != nil {
		return nil, err
	} else if f, err := fallback.New(fps); err != nil {
		return nil, err
	} else if m, err := management.New(r); err != nil {
		return nil, err
	} else {
		result := Lingress{
			RulesRepository: r,
			Proxy:           p,
			Fallback:        f,
			Management:      m,
			http: http.Server{
				Handler: p,
				ErrorLog: support.StdLog(log.Fields{
					"context": "server.http",
				}, log.DebugLevel),
			},
			https: http.Server{
				Handler: p,
				ErrorLog: support.StdLog(log.Fields{
					"context": "server.https",
				}, log.DebugLevel),
			},
			HttpListenAddr:    ":8080",
			HttpsListenAddr:   ":8443",
			MaxHeaderBytes:    2 << 20, // 2MB,
			ReadHeaderTimeout: 30 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       5 * time.Minute,

			AccessLogQueueSize: 5000,
		}

		p.ResultHandler = result.onResult
		p.AccessLogger = result.onAccessLog

		return &result, nil
	}
}

func (instance *Lingress) RegisterFlag(fe support.FlagEnabled, appPrefix string) error {
	fe.Flag("listen.http", "Listen address where the proxy is listening to serve").
		PlaceHolder(instance.HttpListenAddr).
		Envar(support.FlagEnvName(appPrefix, "LISTEN_HTTP")).
		StringVar(&instance.HttpListenAddr)
	fe.Flag("listen.https", "Listen address where the proxy is listening to serve").
		PlaceHolder(instance.HttpsListenAddr).
		Envar(support.FlagEnvName(appPrefix, "LISTEN_HTTPS")).
		StringVar(&instance.HttpsListenAddr)
	fe.Flag("accessLogQueueSize", "Maximum number of accessLog elements that could be queue before blocking.").
		PlaceHolder(fmt.Sprint(instance.AccessLogQueueSize)).
		Envar(support.FlagEnvName(appPrefix, "ACCESS_LOG_QUEUE_SIZE")).
		Uint16Var(&instance.AccessLogQueueSize)
	fe.Flag("client.maxHeaderBytes", "Maximum number of bytes the server will read parsing the request header's keys and values, including the request line. It does not limit the size of the request body.").
		PlaceHolder(fmt.Sprint(instance.MaxHeaderBytes)).
		Envar(support.FlagEnvName(appPrefix, "CLIENT_MAX_HEADER_BYTES")).
		UintVar(&instance.MaxHeaderBytes)
	fe.Flag("client.readHeaderTimeout", "Amount of time allowed to read request headers. The connection's read deadline is reset after reading the headers and the Handler can decide what is considered too slow for the body.").
		PlaceHolder(fmt.Sprint(instance.ReadHeaderTimeout)).
		Envar(support.FlagEnvName(appPrefix, "CLIENT_READ_HEADER_TIMEOUT")).
		DurationVar(&instance.ReadHeaderTimeout)
	fe.Flag("client.writeTimeout", "Maximum duration before timing out writes of the response. It is reset whenever a new request's header is read.").
		PlaceHolder(fmt.Sprint(instance.WriteTimeout)).
		Envar(support.FlagEnvName(appPrefix, "CLIENT_WRITE_TIMEOUT")).
		DurationVar(&instance.WriteTimeout)
	fe.Flag("client.idleTimeout", "Maximum amount of time to wait for the next request when keep-alives are enabled.").
		PlaceHolder(fmt.Sprint(instance.IdleTimeout)).
		Envar(support.FlagEnvName(appPrefix, "CLIENT_IDLE_TIMEOUT")).
		DurationVar(&instance.IdleTimeout)

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
	return nil
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

	tlsConfig, err := instance.createTlsConfig()
	if err != nil {
		return err
	}

	if err := instance.serve(&instance.http, instance.HttpListenAddr, nil, stop); err != nil {
		return err
	}
	if err := instance.serve(&instance.https, instance.HttpsListenAddr, tlsConfig, stop); err != nil {
		return err
	}

	if err := instance.Management.Init(stop); err != nil {
		return err
	}

	return nil
}

func (instance *Lingress) serve(target *http.Server, addr string, tlsConfig *tls.Config, stop support.Channel) error {
	target.Addr = addr
	target.MaxHeaderBytes = int(instance.MaxHeaderBytes)
	target.ReadHeaderTimeout = instance.ReadHeaderTimeout
	target.WriteTimeout = instance.WriteTimeout
	target.IdleTimeout = instance.IdleTimeout

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	ln = tcpKeepAliveListener{ln.(*net.TCPListener)}

	serve := func() error {
		return target.Serve(ln)
	}
	if tlsConfig != nil {
		target.TLSConfig = tlsConfig
		serve = func() error {
			return target.ServeTLS(ln, "", "")
		}
	}

	go func() {
		if err := serve(); err != nil && err != http.ErrServerClosed {
			log.WithError(err).
				WithField("addr", target.Addr).
				Error("server is unable to serve proxy interface")
			stop.Broadcast()
		}
	}()
	log.WithField("addr", target.Addr).
		Info("serve proxy interface")

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
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Minute)
	if err := instance.http.Shutdown(ctx); err != nil {
		log.WithError(err).
			Warnf("cannot graceful shutdown proxy interface %s", instance.http.Addr)
	}
	if err := instance.https.Shutdown(ctx); err != nil {
		log.WithError(err).
			Warnf("cannot graceful shutdown proxy interface %s", instance.http.Addr)
	}
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
