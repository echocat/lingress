package lingress

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	lctx "github.com/echocat/lingress/context"
	"github.com/echocat/lingress/fallback"
	"github.com/echocat/lingress/management"
	"github.com/echocat/lingress/proxy"
	"github.com/echocat/lingress/rules"
	"github.com/echocat/lingress/support"
	"net"
	"net/http"
	"strings"
	"time"
)

type Lingress struct {
	rules              rules.Repository
	proxy              *proxy.Proxy
	fallback           *fallback.Fallback
	management         *management.Management
	server             http.Server
	accessLogQueue     chan accessLogEntry
	accessLogQueueSize uint16
}

type accessLogEntry map[string]interface{}

func New() (*Lingress, error) {
	if r, err := rules.NewRepository(); err != nil {
		return nil, err
	} else if p, err := proxy.New(r); err != nil {
		return nil, err
	} else if f, err := fallback.New(); err != nil {
		return nil, err
	} else if m, err := management.New(r); err != nil {
		return nil, err
	} else {
		result := Lingress{
			rules:      r,
			proxy:      p,
			fallback:   f,
			management: m,
			server: http.Server{
				Addr:              ":8080",
				Handler:           p,
				MaxHeaderBytes:    2 << 20, // 2MB
				ReadHeaderTimeout: 30 * time.Second,
				WriteTimeout:      30 * time.Second,
				IdleTimeout:       5 * time.Minute,
			},
			accessLogQueueSize: 5000,
		}

		p.ResultHandler = result.onResult
		p.AccessLogger = result.onAccessLog

		return &result, nil
	}
}

func (instance *Lingress) RegisterFlag(fe support.FlagEnabled, appPrefix string) error {
	fe.Flag("listen", "Listen address where the proxy is listening to serve").
		PlaceHolder(instance.server.Addr).
		Envar(support.FlagEnvName(appPrefix, "LISTEN")).
		StringVar(&instance.server.Addr)
	fe.Flag("accessLogQueueSize", "Maximum number of accessLog elements that could be queue before blocking.").
		PlaceHolder(fmt.Sprint(instance.accessLogQueueSize)).
		Envar(support.FlagEnvName(appPrefix, "ACCESS_LOG_QUEUE_SIZE")).
		Uint16Var(&instance.accessLogQueueSize)
	fe.Flag("client.maxHeaderBytes", "Maximum number of bytes the server will read parsing the request header's keys and values, including the request line. It does not limit the size of the request body.").
		PlaceHolder(fmt.Sprint(instance.server.MaxHeaderBytes)).
		Envar(support.FlagEnvName(appPrefix, "CLIENT_MAX_HEADER_BYTES")).
		IntVar(&instance.server.MaxHeaderBytes)
	fe.Flag("client.readHeaderTimeout", "Amount of time allowed to read request headers. The connection's read deadline is reset after reading the headers and the Handler can decide what is considered too slow for the body.").
		PlaceHolder(fmt.Sprint(instance.server.ReadHeaderTimeout)).
		Envar(support.FlagEnvName(appPrefix, "CLIENT_READ_HEADER_TIMEOUT")).
		DurationVar(&instance.server.ReadHeaderTimeout)
	fe.Flag("client.writeTimeout", "Maximum duration before timing out writes of the response. It is reset whenever a new request's header is read.").
		PlaceHolder(fmt.Sprint(instance.server.WriteTimeout)).
		Envar(support.FlagEnvName(appPrefix, "CLIENT_WRITE_TIMEOUT")).
		DurationVar(&instance.server.WriteTimeout)
	fe.Flag("client.idleTimeout", "Maximum amount of time to wait for the next request when keep-alives are enabled.").
		PlaceHolder(fmt.Sprint(instance.server.IdleTimeout)).
		Envar(support.FlagEnvName(appPrefix, "CLIENT_IDLE_TIMEOUT")).
		DurationVar(&instance.server.IdleTimeout)

	if err := instance.rules.RegisterFlag(fe, appPrefix); err != nil {
		return err
	}
	if err := instance.proxy.RegisterFlag(fe, appPrefix); err != nil {
		return err
	}
	if err := instance.fallback.RegisterFlag(fe, appPrefix); err != nil {
		return err
	}
	return nil
}

func (instance *Lingress) Init(stopCh chan struct{}) error {
	if err := instance.rules.Init(stopCh); err != nil {
		return err
	}

	if s := instance.accessLogQueueSize; s > 0 {
		queue := make(chan accessLogEntry, instance.accessLogQueueSize)
		go instance.accessLogQueueWorker(stopCh, queue)
		instance.accessLogQueue = queue
	} else {
		instance.accessLogQueue = nil
	}

	go instance.shutdownListener(stopCh)

	ln, err := net.Listen("tcp", instance.server.Addr)
	if err != nil {
		return err
	}
	ln = tcpKeepAliveListener{ln.(*net.TCPListener)}

	go func() {
		if err := instance.server.Serve(ln); err != nil {
			log.WithError(err).
				Fatal("server is unable to serve proxy interface")
		}
	}()

	if err := instance.management.Init(stopCh); err != nil {
		return err
	}

	return nil
}

func (instance *Lingress) shutdownListener(stopCh chan struct{}) {
	<-stopCh
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Minute)
	if err := instance.server.Shutdown(ctx); err != nil {
		log.WithError(err).
			Warn("cannot graceful shutdown proxy interface")
	}
}

func (instance *Lingress) accessLogQueueWorker(stopCh chan struct{}, queue chan accessLogEntry) {
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
	if proceed, err := instance.proxy.Interceptors.Handle(ctx); err != nil {
		ctx.Log().WithError(err).Error("cannot prepare error response")
		return
	} else if !proceed {
		return
	}

	ctx.Stage = lctx.StageSendResponseToClient
	ctx.Client.Status = ctx.Result.Status()
	if ctx.Result == lctx.ResultFailedWithRuleNotFound {
		instance.fallback.Unknown(ctx)
	} else if rr, ok := ctx.Result.(lctx.RedirectResult); ok {
		instance.fallback.Redirect(ctx, ctx.Client.Status, rr.Target)
	} else {
		p := ""
		if u, err := ctx.Client.RequestedUrl(); err == nil && u != nil {
			p = u.Path
		}
		canHandleTemporary := ctx.Client.Request.Method == "GET"
		instance.fallback.Status(ctx, ctx.Client.Status, p, canHandleTemporary)
	}

	ctx.Stage = lctx.StageDone
}
