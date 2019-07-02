package management

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/echocat/lingress/rules"
	"github.com/echocat/lingress/support"
	"net"
	"net/http"
	"strings"
	"time"
)

type Management struct {
	server http.Server
	rules  rules.Repository
}

func New(rulesRepository rules.Repository) (*Management, error) {
	result := &Management{
		rules: rulesRepository,
		server: http.Server{
			Addr:              ":8090",
			MaxHeaderBytes:    2 << 20, // 2MB
			ReadHeaderTimeout: 30 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       5 * time.Minute,
		},
	}
	result.server.Handler = result
	return result, nil

}

func (instance *Management) RegisterFlag(fe support.FlagEnabled, appPrefix string) error {
	fe.Flag("management.listen", "Listen address where the management interface is listening to serve").
		PlaceHolder(instance.server.Addr).
		Envar(support.FlagEnvName(appPrefix, "MANAGEMENT_LISTEN")).
		StringVar(&instance.server.Addr)
	fe.Flag("Management.maxHeaderBytes", "Maximum number of bytes the server will read parsing the request header's keys and values, including the request line. It does not limit the size of the request body.").
		PlaceHolder(fmt.Sprint(instance.server.MaxHeaderBytes)).
		Envar(support.FlagEnvName(appPrefix, "MANAGEMENT_MAX_HEADER_BYTES")).
		IntVar(&instance.server.MaxHeaderBytes)
	fe.Flag("Management.readHeaderTimeout", "Amount of time allowed to read request headers. The connection's read deadline is reset after reading the headers and the Handler can decide what is considered too slow for the body.").
		PlaceHolder(fmt.Sprint(instance.server.ReadHeaderTimeout)).
		Envar(support.FlagEnvName(appPrefix, "MANAGEMENT_READ_HEADER_TIMEOUT")).
		DurationVar(&instance.server.ReadHeaderTimeout)
	fe.Flag("Management.writeTimeout", "Maximum duration before timing out writes of the response. It is reset whenever a new request's header is read.").
		PlaceHolder(fmt.Sprint(instance.server.WriteTimeout)).
		Envar(support.FlagEnvName(appPrefix, "MANAGEMENT_WRITE_TIMEOUT")).
		DurationVar(&instance.server.WriteTimeout)
	fe.Flag("Management.idleTimeout", "Maximum amount of time to wait for the next request when keep-alives are enabled.").
		PlaceHolder(fmt.Sprint(instance.server.IdleTimeout)).
		Envar(support.FlagEnvName(appPrefix, "MANAGEMENT_IDLE_TIMEOUT")).
		DurationVar(&instance.server.IdleTimeout)

	return nil
}

func (instance *Management) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		support.NewGenericResponse(http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed), req).
			StreamJsonTo(resp, req)
		return
	}
	if req.RequestURI == "/health" {
		instance.handleHealth(resp, req)
	} else if req.RequestURI == "/status" {
		instance.handleStatus(resp, req)
	} else if req.RequestURI == "/rules" {
		instance.handleRules(resp, req, "")
	} else if strings.HasPrefix(req.RequestURI, "/rules/") {
		instance.handleRules(resp, req, req.RequestURI[7:])
	} else {
		support.NewGenericResponse(http.StatusNotFound, http.StatusText(http.StatusNotFound), req).
			StreamJsonTo(resp, req)
	}
}

func (instance *Management) handleHealth(resp http.ResponseWriter, req *http.Request) {
	support.NewGenericResponse(http.StatusOK, http.StatusText(http.StatusOK), req).
		StreamJsonTo(resp, req)
}

func (instance *Management) handleStatus(resp http.ResponseWriter, req *http.Request) {
	var numberOfRules uint
	var numberOfRequests uint64
	var totalDuration time.Duration
	_ = instance.rules.All(func(rule rules.Rule) error {
		numberOfRules++
		numberOfRequests += rule.Statistics().NumberOfUsages()
		totalDuration += rule.Statistics().TotalDuration()
		return nil
	})
	runtime := support.Runtime()

	data := map[string]interface{}{
		"runtime": map[string]interface{}{
			"groupId":    runtime.GroupId,
			"artifactId": runtime.ArtifactId,
			"revision":   runtime.Revision,
			"branch":     runtime.Branch,
			"build":      runtime.Build,
			"goVersion":  runtime.GoVersion,
			"os":         runtime.Os,
			"arch":       runtime.Arch,
		},
		"statistics": map[string]interface{}{
			"numberOfRules":    numberOfRules,
			"numberOfRequests": numberOfRequests,
			"totalDuration":    totalDuration / time.Microsecond,
		},
	}
	support.NewGenericResponse(http.StatusOK, http.StatusText(http.StatusOK), req).
		WithData(data).
		StreamJsonTo(resp, req)
}

func (instance *Management) handleRules(resp http.ResponseWriter, req *http.Request, requestedSource string) {
	result := make(map[string][]map[string]interface{})

	if err := instance.rules.All(func(rule rules.Rule) error {
		source := rule.Source().String()
		if requestedSource == "" || requestedSource == source {
			entry := map[string]interface{}{
				"statistics": rule.Statistics(),
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
		log.WithError(err).
			Error("unable to read rules")
		support.NewGenericResponse(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError), req).
			StreamJsonTo(resp, req)
		return
	}

	var data interface{} = result

	if requestedSource != "" {
		var ok bool
		if data, ok = result[requestedSource]; !ok {
			support.NewGenericResponse(http.StatusNotFound, http.StatusText(http.StatusNotFound), req).
				StreamJsonTo(resp, req)
			return
		}
	}

	support.NewGenericResponse(http.StatusOK, http.StatusText(http.StatusOK), req).
		WithData(data).
		StreamJsonTo(resp, req)
}

func (instance *Management) Init(stopCh chan struct{}) error {
	go instance.shutdownListener(stopCh)

	ln, err := net.Listen("tcp", instance.server.Addr)
	if err != nil {
		return err
	}

	go func() {
		if err := instance.server.Serve(ln); err != nil {
			log.WithError(err).
				Fatal("server is unable to serve management interface")
		}
	}()

	return nil
}

func (instance *Management) shutdownListener(stopCh chan struct{}) {
	<-stopCh
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Minute)
	if err := instance.server.Shutdown(ctx); err != nil {
		log.WithError(err).
			Warn("cannot graceful shutdown management interface")
	}
}
