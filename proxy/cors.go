package proxy

import (
	"fmt"
	"github.com/echocat/lingress/context"
	"github.com/echocat/lingress/rules"
	"github.com/echocat/lingress/value"
	"net"
	"net/http"
	"strconv"
	"time"
)

func init() {
	DefaultInterceptors.Add(NewCorsInterceptor())
}

var (
	defaultMaxAge = value.NewDuration(24 * time.Hour)
)

type CorsInterceptor struct{}

func NewCorsInterceptor() *CorsInterceptor {
	return &CorsInterceptor{}
}

func (this *CorsInterceptor) Name() string {
	return "cors"
}

func (this *CorsInterceptor) HandlesStages() []context.Stage {
	return []context.Stage{context.StageEvaluateClientRequest, context.StagePrepareClientResponse}
}

func (this *CorsInterceptor) Handle(ctx *context.Context) (proceed bool, err error) {
	switch ctx.Stage {
	case context.StageEvaluateClientRequest:
		return this.enableCors(ctx)
	case context.StagePrepareClientResponse:
		return this.forceCorsHeaders(ctx)
	}
	return true, nil
}

func (this *CorsInterceptor) enableCors(ctx *context.Context) (proceed bool, err error) {
	optionsCors := rules.OptionsCorsOf(ctx.Rule)
	if !ctx.Settings.Cors.Enabled.Select(optionsCors.Enabled).GetOr(false) {
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
	if !ctx.Settings.Cors.AllowedOriginsHost.Evaluate(optionsCors.AllowedOriginsHost).Matches(host) {
		ctx.Client.Response.Header().Set("X-Cors-Hint", "origin-ignored")
		return true, nil
	}

	if ctx.Client.Request.Method == "OPTIONS" {
		ctx.Client.Response.Header().Set("X-Reason", "cors-options")
		ctx.Result = context.ResultOk
		return false, nil
	}

	return true, nil
}

func (this *CorsInterceptor) deleteHeaders(h http.Header) {
	h.Del("Access-Control-Allow-Origin")
	h.Del("Access-Control-Allow-Credentials")
	h.Del("Access-Control-Allow-Methods")
	h.Del("Access-Control-Allow-Headers")
	h.Del("Access-Control-Max-Age")
}

func (this *CorsInterceptor) forceCorsHeaders(ctx *context.Context) (proceed bool, err error) {
	proceed = true
	cors := rules.OptionsCorsOf(ctx.Rule)
	enabled := ctx.Settings.Cors.Enabled.Select(cors.Enabled)
	h := ctx.Client.Response.Header()

	if enabled.GetOr(false) {
		origin, oErr := ctx.Client.Origin()
		if oErr != nil {
			return false, fmt.Errorf("cannot enable cors: %w", oErr)
		}
		if origin == nil {
			this.deleteHeaders(h)
			return
		}
		host := origin.Host
		if sHost, _, err := net.SplitHostPort(host); err == nil {
			host = sHost
		}
		if !ctx.Settings.Cors.AllowedOriginsHost.Evaluate(cors.AllowedOriginsHost).Matches(host) {
			this.deleteHeaders(h)
			return
		}
		h.Set("Access-Control-Allow-Origin", origin.String())
		h.Set("Access-Control-Allow-Credentials", fmt.Sprint(ctx.Settings.Cors.AllowedCredentials.Evaluate(cors.AllowedCredentials).GetOr(true)))
		h.Set("Access-Control-Allow-Methods", ctx.Settings.Cors.AllowedMethods.Evaluate(cors.AllowedMethods).String())
		h.Set("Access-Control-Allow-Headers", ctx.Settings.Cors.AllowedHeaders.Evaluate(cors.AllowedHeaders).String())
		h.Set("Access-Control-Max-Age", strconv.Itoa(ctx.Settings.Cors.MaxAge.EvaluateOr(cors.MaxAge, defaultMaxAge).AsSeconds()))
	} else if enabled.IsForced() {
		this.deleteHeaders(h)
	}

	return
}
