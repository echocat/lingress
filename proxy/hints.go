package proxy

import (
	"github.com/echocat/lingress/context"
	"net"
)

func init() {
	DefaultInterceptors.AddFunc("upstreamHints", UpstreamHintsInterceptor, context.StagePrepareUpstreamRequest)
	DefaultInterceptors.AddFunc("clientHints", ClientHintsInterceptor, context.StagePrepareClientResponse)
}

func UpstreamHintsInterceptor(ctx *context.Context) (proceed bool, err error) {
	h := ctx.Upstream.Request.Header
	h.Set("X-Source", ctx.Rule.Source().String())
	h.Set("X-Request-Id", ctx.Id.String())
	h.Set("X-Correlation-Id", ctx.CorrelationId.String())

	if u, err := ctx.Client.RequestedUrl(); err != nil {
		return false, err
	} else if u != nil {
		ctx.Upstream.Request.Host = u.Host
		h.Set("Host", u.Host)
		h.Set("X-Forwarded-Host", u.Host)
		h.Set("X-Forwarded-Proto", u.Scheme)
		h.Set("X-Original-Uri", u.RequestURI())
	}

	if r, _, err := net.SplitHostPort(ctx.Client.Request.RemoteAddr); err != nil {
		h.Add("X-Forwarded-For", ctx.Client.Request.RemoteAddr)
	} else {
		h.Add("X-Forwarded-For", r)
	}

	return true, nil
}

func ClientHintsInterceptor(ctx *context.Context) (proceed bool, err error) {
	h := ctx.Client.Response.Header()
	if r := ctx.Rule; r != nil {
		h.Set("X-Source", r.Source().String())
	}
	h.Set("X-Request-Id", ctx.Id.String())
	h.Set("X-Correlation-Id", ctx.CorrelationId.String())
	return true, nil
}
