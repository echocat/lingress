package proxy

import (
	"github.com/echocat/lingress/context"
	"github.com/echocat/lingress/rules"
	"strings"
)

func init() {
	DefaultInterceptors.AddFunc("prefix", PrefixInterceptor, context.StagePrepareUpstreamRequest)
}

func PrefixInterceptor(ctx *context.Context) (proceed bool, err error) {
	r := ctx.Rule
	u := ctx.Upstream.Request.URL
	opts := rules.OptionsPrefixOf(r.Options())

	if len(opts.PathPrefix) > 0 || opts.StripRulePathPrefix.GetOr(false) {
		path, err := rules.ParsePath(u.Path, true)
		if err != nil {
			return false, err
		}
		if opts.StripRulePathPrefix.GetOr(false) {
			path = stripPrefix(path, r.Path())
		}
		if len(opts.PathPrefix) > 0 {
			path = append(opts.PathPrefix, path...)
		}
		u.Path = "/" + strings.Join(path, "/")
	}

	if opts.XForwardedPrefix.GetOr(false) {
		prefix := ""
		if ctx.Client.FromOtherReverseProxy {
			prefix = ctx.Client.Request.Header.Get("X-Forwarded-Prefix")
		}
		prefix += "/" + strings.Join(r.Path(), "/")
		ctx.Upstream.Request.Header.Set("X-Forwarded-Prefix", prefix)
	} else {
		ctx.Upstream.Request.Header.Del("X-Forwarded-Prefix")
	}

	return true, nil
}
