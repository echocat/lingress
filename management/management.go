package management

import (
	"context"
	lctx "github.com/echocat/lingress/context"
	"github.com/echocat/lingress/rules"
	"github.com/echocat/lingress/server"
	"github.com/echocat/lingress/settings"
	"github.com/echocat/lingress/support"
	"github.com/echocat/slf4g"
	"github.com/echocat/slf4g/level"
	sdk "github.com/echocat/slf4g/sdk/bridge"
	"net"
	"net/http"
	"net/http/pprof"
	"strings"
	"time"
)

type Management struct {
	settings *settings.Settings

	Metrics *Metrics
	Logger  log.Logger

	server http.Server
	rules  rules.Repository
}

func New(s *settings.Settings, connectorIds []server.ConnectorId, rulesRepository rules.Repository, logger log.Logger) (*Management, error) {
	result := &Management{
		settings: s,
		Metrics:  NewMetrics(connectorIds, rulesRepository),
		Logger:   logger,
		rules:    rulesRepository,
		server: http.Server{
			ErrorLog: sdk.NewWrapper(logger, level.Debug),
		},
	}
	result.server.Handler = result
	return result, nil

}

func (this *Management) CollectContext(ctx *lctx.Context) {
	this.Metrics.CollectContext(ctx)
}

func (this *Management) CollectClientStarted(connector server.ConnectorId) func() {
	return this.Metrics.CollectClientStarted(connector)
}

func (this *Management) CollectUpstreamStarted() func() {
	return this.Metrics.CollectUpstreamStarted()
}

func (this *Management) getLogger() log.Logger {
	return this.Logger
}

func (this *Management) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		support.NewGenericResponse(http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed), req).
			StreamJsonTo(resp, req, this.getLogger)
		return
	}
	isPprof := this.settings.Management.Pprof.Get()
	if req.URL.Path == "/health" {
		this.handleHealth(resp, req)
	} else if req.URL.Path == "/status" {
		this.handleStatus(resp, req)
	} else if req.URL.Path == "/metrics" {
		this.handleMetrics(resp, req)
	} else if req.URL.Path == "/rules" {
		this.handleRules(resp, req, "")
	} else if strings.HasPrefix(req.URL.Path, "/rules/") {
		this.handleRules(resp, req, req.URL.Path[7:])
	} else if isPprof && strings.HasPrefix(req.URL.Path, "/debug/pprof/cmdline") {
		pprof.Cmdline(resp, req)
	} else if isPprof && strings.HasPrefix(req.URL.Path, "/debug/pprof/profile") {
		pprof.Profile(resp, req)
	} else if isPprof && strings.HasPrefix(req.URL.Path, "/debug/pprof/symbol") {
		pprof.Symbol(resp, req)
	} else if isPprof && strings.HasPrefix(req.URL.Path, "/debug/pprof/trace") {
		pprof.Trace(resp, req)
	} else if isPprof && strings.HasPrefix(req.URL.Path, "/debug/pprof/") {
		pprof.Index(resp, req)
	} else if isPprof && req.URL.Path == "/debug/pprof" {
		http.Redirect(resp, req, "/debug/pprof/", http.StatusMovedPermanently)
	} else {
		support.NewGenericResponse(http.StatusNotFound, http.StatusText(http.StatusNotFound), req).
			StreamJsonTo(resp, req, this.getLogger)
	}
}

func (this *Management) handleHealth(resp http.ResponseWriter, req *http.Request) {
	support.NewGenericResponse(http.StatusOK, http.StatusText(http.StatusOK), req).
		StreamJsonTo(resp, req, this.getLogger)
}

func (this *Management) handleStatus(resp http.ResponseWriter, req *http.Request) {
	var numberOfRules uint
	var numberOfRequests uint64
	var totalDuration time.Duration
	_ = this.rules.All(func(rule rules.Rule) error {
		numberOfRules++
		numberOfRequests += rule.Statistics().NumberOfUsages()
		totalDuration += rule.Statistics().TotalDuration()
		return nil
	})
	runtime := support.Runtime()

	data := map[string]interface{}{
		"runtime": map[string]interface{}{
			"revision":  runtime.Revision,
			"branch":    runtime.Branch,
			"build":     runtime.Build,
			"goVersion": runtime.GoVersion,
			"os":        runtime.Os,
			"arch":      runtime.Arch,
		},
		"statistics": map[string]interface{}{
			"numberOfRules":    numberOfRules,
			"numberOfRequests": numberOfRequests,
			"totalDuration":    totalDuration / time.Microsecond,
		},
	}
	support.NewGenericResponse(http.StatusOK, http.StatusText(http.StatusOK), req).
		WithData(data).
		StreamJsonTo(resp, req, this.getLogger)
}

func (this *Management) handleMetrics(resp http.ResponseWriter, req *http.Request) {
	this.Metrics.Handler.ServeHTTP(resp, req)
}

func (this *Management) handleRules(resp http.ResponseWriter, req *http.Request, requestedSource string) {
	result := make(map[string][]map[string]interface{})

	if err := this.rules.All(func(rule rules.Rule) error {
		source := rule.Source().String()
		if requestedSource == "" || requestedSource == source {
			entry := map[string]interface{}{
				"statistics": rule.Statistics(),
				"pathType":   rule.PathType(),
			}
			if h := rule.Host(); h != "" {
				entry["host"] = h
			}
			if p := rule.Path(); p != nil {
				entry["path"] = "/" + strings.Join(p, "/")
			}
			if o := rule.Options(); o.IsRelevant() {
				entry["options"] = o
			}
			if b := rule.Backend(); b != nil {
				entry["backend"] = b.String()
			}

			if entries, ok := result[source]; ok {
				result[source] = append(entries, entry)
			} else {
				result[source] = []map[string]interface{}{entry}
			}
		}
		return nil
	}); err != nil {
		this.Logger.
			WithError(err).
			Error("Unable to read rules.")
		support.NewGenericResponse(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), req).
			StreamJsonTo(resp, req, this.getLogger)
		return
	}

	var data interface{} = result

	if requestedSource != "" {
		var ok bool
		if data, ok = result[requestedSource]; !ok {
			support.NewGenericResponse(http.StatusNotFound, http.StatusText(http.StatusNotFound), req).
				StreamJsonTo(resp, req, this.getLogger)
			return
		}
	}

	support.NewGenericResponse(http.StatusOK, http.StatusText(http.StatusOK), req).
		WithData(data).
		StreamJsonTo(resp, req, this.getLogger)
}

func (this *Management) Init(stop support.Channel) error {
	go this.shutdownListener(stop)

	if err := this.settings.Management.ApplyToHttpServer(&this.server); err != nil {
		return err
	}

	ln, err := net.Listen("tcp", this.server.Addr)
	if err != nil {
		return err
	}

	if this.settings.Management.Pprof.Get() {
		this.Logger.
			With("addr", this.server.Addr).
			Warn("DO NOT USE IN PRODUCTION!" +
				" pprof endpoints are activated for debugging at listening." +
				" This functionality is only for debug purposes.",
			)
	}

	go func() {
		if err := this.server.Serve(ln); err != nil && err != http.ErrServerClosed {
			this.Logger.
				WithError(err).
				With("addr", this.server.Addr).
				Error("Server is unable to serve management interface.")
			stop.Broadcast()
		}
	}()
	this.Logger.
		With("addr", this.server.Addr).
		Info("Serve management interface...")

	return nil
}

func (this *Management) shutdownListener(stop support.Channel) {
	stop.Wait()
	ctx, cancelFnc := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancelFnc()
	if err := this.server.Shutdown(ctx); err != nil {
		this.Logger.
			WithError(err).
			With("addr", this.server.Addr).
			Warn("Cannot graceful shutdown management interface.")
	}
}
