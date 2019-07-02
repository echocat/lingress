package proxy

import (
	"fmt"
	"github.com/echocat/lingress/context"
	"github.com/echocat/lingress/rules"
	"github.com/echocat/lingress/support"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func init() {
	DefaultInterceptors.Add(NewCorsInterceptor())
}

type CorsInterceptor struct {
	AllowedOriginsHost HostPatterns
	AllowedMethods     string
	AllowedHeaders     string
	AllowedCredentials bool
	MaxAge             time.Duration
	Enabled            rules.ForceableBool
}

func NewCorsInterceptor() *CorsInterceptor {
	return &CorsInterceptor{
		AllowedOriginsHost: HostPatterns{},
		AllowedMethods: strings.Join([]string{
			http.MethodGet,
			http.MethodHead,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodConnect,
			http.MethodOptions,
			http.MethodTrace,
			http.MethodPatch,
		}, ","),
		AllowedHeaders:     "*",
		AllowedCredentials: true,
		MaxAge:             24 * time.Hour,
		Enabled:            rules.ForceableBool{Value: false, Forced: false},
	}
}

func (instance *CorsInterceptor) Name() string {
	return "cors"
}

func (instance *CorsInterceptor) HandlesStages() []context.Stage {
	return []context.Stage{context.StageEvaluateClientRequest, context.StagePrepareClientResponse}
}

func (instance *CorsInterceptor) Handle(ctx *context.Context) (proceed bool, err error) {
	switch ctx.Stage {
	case context.StageEvaluateClientRequest:
		return instance.enableCors(ctx)
	case context.StagePrepareClientResponse:
		return instance.forceCorsHost(ctx)
	}
	return true, nil
}

func (instance *CorsInterceptor) RegisterFlag(fe support.FlagEnabled, appPrefix string) error {
	fe.Flag("cors.enabled", "If set if will be used if annotation lingress.echocat.org/cors is absent. If this value is prefix with ! it overrides everything regardless what was set in the annotation.").
		PlaceHolder(fmt.Sprint(instance.Enabled)).
		Envar(support.FlagEnvName(appPrefix, "CORS_ENABLED")).
		SetValue(&instance.Enabled)
	fe.Flag("cors.allowedOriginHosts", "Host patterns which are allowed to be referenced by Origin headers. '*' means all.").
		PlaceHolder(instance.AllowedOriginsHost.String()).
		Envar(support.FlagEnvName(appPrefix, "CORS_ALLOWED_ORIGIN_HOSTS")).
		SetValue(&instance.AllowedOriginsHost)
	fe.Flag("cors.allowedMethods", "Methods which are allowed to be referenced.").
		PlaceHolder(fmt.Sprint(instance.AllowedMethods)).
		Envar(support.FlagEnvName(appPrefix, "CORS_ALLOWED_METHODS")).
		StringVar(&instance.AllowedMethods)
	fe.Flag("cors.allowedHeaders", "Headers which are allowed to be referenced. '*' means all.").
		PlaceHolder(fmt.Sprint(instance.AllowedHeaders)).
		Envar(support.FlagEnvName(appPrefix, "CORS_ALLOWED_HEADERS")).
		StringVar(&instance.AllowedHeaders)
	fe.Flag("cors.allowedCredentials", "Credentials are allowed.").
		PlaceHolder(fmt.Sprint(instance.AllowedCredentials)).
		Envar(support.FlagEnvName(appPrefix, "CORS_ALLOWED_CREDENTIALS")).
		BoolVar(&instance.AllowedCredentials)
	fe.Flag("cors.maxAge", "How long the response to the preflight request can be cached for without sending another preflight request").
		PlaceHolder(fmt.Sprint(instance.MaxAge)).
		Envar(support.FlagEnvName(appPrefix, "CORS_MAX_AGE")).
		DurationVar(&instance.MaxAge)
	return nil
}

func (instance *CorsInterceptor) enableCors(ctx *context.Context) (proceed bool, err error) {
	if !ctx.Rule.Options().Cors.IsEnabledOrForced(instance.Enabled) {
		return true, nil
	}
	origin, err := ctx.Client.Origin()
	if err != nil {
		return false, fmt.Errorf("cannot enable cors: %v", err)
	}
	if origin == nil {
		return true, nil
	}
	if !instance.AllowedOriginsHost.Matches(origin.Host) {
		ctx.Client.Response.Header().Set("X-Reason", "cors-origin-forbidden")
		ctx.Result = context.ResultFailedWithAccessDenied
		return false, nil
	}
	h := ctx.Client.Response.Header()
	h.Set("Access-Control-Allow-Origin", origin.String())
	h.Set("Access-Control-Allow-Credentials", fmt.Sprint(instance.AllowedCredentials))
	h.Set("Access-Control-Allow-Methods", instance.AllowedMethods)
	h.Set("Access-Control-Allow-Headers", instance.AllowedHeaders)
	h.Set("Access-Control-Max-Age", strconv.Itoa(int(instance.MaxAge/time.Second)))

	if ctx.Client.Request.Method == "OPTIONS" {
		ctx.Client.Response.Header().Set("X-Reason", "cors-options")
		ctx.Result = context.ResultOk
		return false, nil
	}

	return true, nil
}

func (instance *CorsInterceptor) forceCorsHost(ctx *context.Context) (proceed bool, err error) {
	h := ctx.Client.Response.Header()
	if acao := h.Get("Access-Control-Allow-Origin"); acao != "" {
		origin, err := ctx.Client.Origin()
		if err != nil {
			return false, fmt.Errorf("cannot enable cors: %v", err)
		}
		if origin == nil || !instance.AllowedOriginsHost.Matches(origin.Host) {
			h.Del("Access-Control-Allow-Origin")
			return true, nil
		}
		h.Set("Access-Control-Allow-Origin", origin.String())
	}
	return true, nil
}

type HostPatterns struct {
	Values [][]string
}

func (instance HostPatterns) Matches(test string) bool {
	if instance.Values == nil || len(instance.Values) == 0 {
		return true
	}

	testParts := strings.Split(test, ".")

	for _, candidateParts := range instance.Values {
		if len(testParts) == len(candidateParts) {
			allMatches := true
			for i, cp := range candidateParts {
				if cp != "*" && cp != testParts[i] {
					allMatches = false
					break
				}
			}
			if allMatches {
				return true
			}
		}
	}

	return false
}

func (instance *HostPatterns) String() string {
	if instance.Values == nil || len(instance.Values) == 0 {
		return "*"
	}

	parts := make([]string, len(instance.Values))
	for i, part := range instance.Values {
		parts[i] = strings.Join(part, ".")
	}
	return strings.Join(parts, ",")
}

func (instance *HostPatterns) Set(in string) error {
	if in == "*" || in == "" {
		instance.Values = nil
		return nil
	}

	var nv [][]string

	for _, host := range strings.Split(in, ",") {
		host = strings.TrimSpace(host)
		if host != "" {
			hostParts := strings.Split(host, ".")
			nv = append(nv, hostParts)
		}
	}

	instance.Values = nv
	return nil
}
