package proxy

import (
	"fmt"
	"github.com/echocat/lingress/context"
	"github.com/echocat/lingress/rules"
	"github.com/echocat/lingress/support"
	"net"
	"net/http"
	"strconv"
	"time"
)

func init() {
	DefaultInterceptors.Add(NewCorsInterceptor())
}

var (
	defaultMaxAge = rules.Duration(24 * time.Hour)
)

type CorsInterceptor struct {
	AllowedOriginsHost rules.ForceableHostPatterns
	AllowedMethods     rules.ForceableMethods
	AllowedHeaders     rules.ForceableHeaderNames
	AllowedCredentials rules.ForceableBool
	MaxAge             rules.ForceableDuration
	Enabled            rules.ForceableBool
}

func NewCorsInterceptor() *CorsInterceptor {
	return &CorsInterceptor{
		AllowedOriginsHost: rules.NewForceableHostPatterns(rules.HostPatterns{}, false),
		AllowedMethods:     rules.NewForceableMethods(rules.Methods{}, false),
		AllowedHeaders:     rules.NewForceableHeaders(rules.HeaderNames{}, false),
		AllowedCredentials: rules.NewForceableBool(rules.True, false),
		MaxAge:             rules.NewForceableDuration(defaultMaxAge, false),
		Enabled:            rules.NewForceableBool(rules.False, false),
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
		return instance.forceCorsHeaders(ctx)
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
		SetValue(&instance.AllowedMethods)
	fe.Flag("cors.allowedHeaders", "Headers which are allowed to be referenced. '*' means all.").
		PlaceHolder(fmt.Sprint(instance.AllowedHeaders)).
		Envar(support.FlagEnvName(appPrefix, "CORS_ALLOWED_HEADERS")).
		SetValue(&instance.AllowedHeaders)
	fe.Flag("cors.allowedCredentials", "Credentials are allowed.").
		PlaceHolder(fmt.Sprint(instance.AllowedCredentials)).
		Envar(support.FlagEnvName(appPrefix, "CORS_ALLOWED_CREDENTIALS")).
		SetValue(&instance.AllowedCredentials)
	fe.Flag("cors.maxAge", "How long the response to the preflight request can be cached for without sending another preflight request").
		PlaceHolder(fmt.Sprint(instance.MaxAge)).
		Envar(support.FlagEnvName(appPrefix, "CORS_MAX_AGE")).
		SetValue(&instance.MaxAge)
	return nil
}

func (instance *CorsInterceptor) enableCors(ctx *context.Context) (proceed bool, err error) {
	optionsCors := rules.OptionsCorsOf(ctx.Rule)
	if !instance.Enabled.Select(optionsCors.Enabled).GetOr(false) {
		return true, nil
	}
	origin, err := ctx.Client.Origin()
	if err != nil {
		return false, fmt.Errorf("cannot enable cors: %v", err)
	}
	if origin == nil {
		return true, nil
	}
	host := origin.Host
	if sHost, _, err := net.SplitHostPort(host); err == nil {
		host = sHost
	}
	if !instance.AllowedOriginsHost.Evaluate(optionsCors.AllowedOriginsHost, nil).Matches(host) {
		ctx.Client.Response.Header().Set("X-Reason", "cors-origin-forbidden")
		ctx.Result = context.ResultFailedWithAccessDenied
		return false, nil
	}

	if ctx.Client.Request.Method == "OPTIONS" {
		ctx.Client.Response.Header().Set("X-Reason", "cors-options")
		ctx.Result = context.ResultOk
		return false, nil
	}

	return true, nil
}

func (instance *CorsInterceptor) deleteHeaders(h http.Header) {
	h.Del("Access-Control-Allow-Origin")
	h.Del("Access-Control-Allow-Credentials")
	h.Del("Access-Control-Allow-Methods")
	h.Del("Access-Control-Allow-Headers")
	h.Del("Access-Control-Max-Age")
}

func (instance *CorsInterceptor) forceCorsHeaders(ctx *context.Context) (proceed bool, err error) {
	proceed = true
	cors := rules.OptionsCorsOf(ctx.Rule)
	enabled := instance.Enabled.Select(cors.Enabled)
	h := ctx.Client.Response.Header()

	if enabled.GetOr(false) {
		origin, oErr := ctx.Client.Origin()
		if oErr != nil {
			return false, fmt.Errorf("cannot enable cors: %w", oErr)
		}
		if origin == nil || !instance.AllowedOriginsHost.Evaluate(cors.AllowedOriginsHost, nil).Matches(origin.Host) {
			instance.deleteHeaders(h)
			return
		}
		h.Set("Access-Control-Allow-Origin", origin.String())
		h.Set("Access-Control-Allow-Credentials", fmt.Sprint(instance.AllowedCredentials.Evaluate(cors.AllowedCredentials, true)))
		h.Set("Access-Control-Allow-Methods", instance.AllowedMethods.Evaluate(cors.AllowedMethods, nil).String())
		h.Set("Access-Control-Allow-Headers", instance.AllowedHeaders.Evaluate(cors.AllowedHeaders, nil).String())
		h.Set("Access-Control-Max-Age", strconv.Itoa(instance.MaxAge.Evaluate(cors.MaxAge, defaultMaxAge).AsSeconds()))
	} else if enabled.IsForced() {
		instance.deleteHeaders(h)
	}

	return
}
