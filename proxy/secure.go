package proxy

import (
	"github.com/echocat/lingress/context"
	"github.com/echocat/lingress/rules"
	"net"
	"net/http"
)

func init() {
	DefaultInterceptors.Add(&ForceSecureInterceptor{})
	DefaultInterceptors.AddFunc("whitelistedRemotes", WhitelistedRemotesInterceptor, context.StageEvaluateClientRequest)
	DefaultInterceptors.AddFunc("removeServerHeader", RemoveServerHeader, context.StagePrepareClientResponse)
}

type ForceSecureInterceptor struct{}

func (this *ForceSecureInterceptor) Name() string {
	return "forceSecure"
}

func (this *ForceSecureInterceptor) HandlesStages() []context.Stage {
	return []context.Stage{context.StageEvaluateClientRequest}
}

func (this *ForceSecureInterceptor) Handle(ctx *context.Context) (proceed bool, err error) {
	opts := rules.OptionsSecureOf(ctx.Rule)
	if !ctx.Settings.Tls.Forced.Evaluate(opts.ForceSecure).GetOr(true) {
		return true, nil
	}

	req := ctx.Client.Request
	resp := ctx.Client.Response
	if req == nil || resp == nil {
		return true, nil
	}

	u, err := ctx.Client.RequestedUrl()
	if err != nil || u == nil {
		return true, err
	}

	if u.Scheme == "https" {
		return true, nil
	}

	cu := *u
	u = &cu
	u.Scheme = "https"

	var status int
	switch ctx.Client.Request.Method {
	case "GET", "HEAD", "CONNECT", "OPTIONS", "TRACE":
		status = http.StatusMovedPermanently
	default:
		status = http.StatusPermanentRedirect
	}

	ctx.Client.Response.Header().Set("X-Reason", "force-secure")
	ctx.Result = context.RedirectResult{
		StatusCode: status,
		Target:     u.String(),
	}

	return false, nil
}

func WhitelistedRemotesInterceptor(ctx *context.Context) (proceed bool, err error) {
	r := ctx.Rule
	opts := rules.OptionsSecureOf(r)
	if r == nil {
		return true, nil
	}
	wr := opts.WhitelistedRemotes
	if wr == nil || len(wr) <= 0 {
		return true, nil
	}

	address, err := ctx.Client.Address()
	if err != nil {
		return false, err
	}

	ips, err := net.LookupIP(address)
	if err != nil {
		return false, err
	}

	for _, ip := range ips {
		for _, candidate := range wr {
			if candidate.Matches(ip) {
				return true, nil
			}
		}
	}

	ctx.Client.Response.Header().Set("X-Reason", "not-whitelisted")
	ctx.Result = context.ResultFailedWithAccessDenied
	return false, nil
}

func RemoveServerHeader(ctx *context.Context) (proceed bool, err error) {
	ctx.Client.Response.Header().Del("Server")
	return true, nil
}
