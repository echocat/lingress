package fallback

import (
	"github.com/gobuffalo/packr"
	"github.com/valyala/fasthttp"
	"github.com/echocat/lingress/context"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strings"
)

var (
	staticFiles = packr.NewBox("../resources/static")
)

// TODO! We should move this maybe better to configurations rather than in this code

func init() {
	_ = mime.AddExtensionType(".ico", "image/x-icon")
	_ = mime.AddExtensionType(".yml", "application/x-yaml")
	_ = mime.AddExtensionType(".yaml", "application/x-yaml")
	_ = mime.AddExtensionType(".json", "application/json")
	_ = mime.AddExtensionType(".txt", "text/plain; charset=utf-8")
}

func (instance *Fallback) Unknown(ctx *context.Context) {
	u, err := ctx.Client.RequestedUrl()
	if err != nil {
		ctx.Log().WithError(err).Error("cannot resolve requestedUrl for handling fallback")
		http.Error(ctx.Client.Response, "Not found", ctx.Client.Status)
		return
	}
	if u == nil {
		ctx.Log().Error("cannot resolve requestedUrl for handling fallback")
		http.Error(ctx.Client.Response, "Not found", ctx.Client.Status)
		return
	}

	p := u.Path

	if p == "/" || p == "" {
		instance.index(ctx, u)
	} else if p == "/view" {
		instance.view(ctx, u)
	} else if p == "/app" || p == "/app/" {
		instance.app(ctx, u)
	} else if p == "/health" {
		instance.health(ctx, u)
	} else {
		instance.fallback(ctx, u)
	}
}

func (instance *Fallback) index(ctx *context.Context, u *url.URL) {
	if !instance.isValidHostForFixing(ctx, u) {
		instance.fallback(ctx, u)
		return
	}
	ctx.Result = context.ResultSuccess
	instance.Redirect(ctx, http.StatusTemporaryRedirect, instance.constructNewTarget(ctx, "", u))
}

func (instance *Fallback) view(ctx *context.Context, u *url.URL) {
	if !instance.isValidHostForFixing(ctx, u) {
		instance.fallback(ctx, u)
		return
	}
	ctx.Result = context.ResultSuccess
	instance.Redirect(ctx, http.StatusPermanentRedirect, instance.constructNewTarget(ctx, "", u))
}

func (instance *Fallback) app(ctx *context.Context, u *url.URL) {
	if !instance.isValidHostForFixing(ctx, u) {
		instance.fallback(ctx, u)
		return
	}
	pathSuffix := ""
	if strings.HasPrefix(u.Path, "/app/") {
		pathSuffix = u.Path[5:]
	}
	ctx.Result = context.ResultSuccess
	instance.Redirect(ctx, http.StatusPermanentRedirect, instance.constructNewTarget(ctx, pathSuffix, u))
}

func (instance *Fallback) health(ctx *context.Context, u *url.URL) {
	ctx.Result = context.ResultSuccess
	ctx.Client.Response.Header().Set("Content-Type", "application/json")
	ctx.Client.Status = http.StatusOK
	ctx.Client.Response.WriteHeader(ctx.Client.Status)
	_, _ = ctx.Client.Response.Write([]byte(`{"status":"UP"}`))
}

func (instance *Fallback) constructNewTarget(ctx *context.Context, pathSuffix string, u *url.URL) string {
	result := *u
	full := false
	if instance.isFixToHttpsRequired(ctx, u) {
		result.Scheme = "https"
		full = true
	}
	result.Path = instance.fixTarget
	if pathSuffix != "" {
		if pathSuffix[0] == '/' {
			result.Path += pathSuffix[1:]
		} else {
			result.Path += pathSuffix
		}
	}

	if full {
		return result.String()
	} else {
		return result.RequestURI()
	}
}

func (instance *Fallback) isValidHostForFixing(ctx *context.Context, u *url.URL) bool {
	if instance.fixTargetForHostsPattern == nil {
		return false
	}
	return instance.fixTargetForHostsPattern.MatchString(u.Host)
}

func (instance *Fallback) isFixToHttpsRequired(ctx *context.Context, u *url.URL) bool {
	if !instance.fixTargetToHttps {
		return false
	}
	return u.Scheme != "https"
}

func (instance *Fallback) fallback(ctx *context.Context, u *url.URL) {
	if p := ctx.Client.Request.RequestURI; !path.IsAbs(p) || path.Clean(p) != p {
		instance.Status(ctx, fasthttp.StatusNotFound, p, false)
	} else if f, err := staticFiles.Open(p); err != nil {
		instance.Status(ctx, fasthttp.StatusNotFound, p, false)
	} else {
		//noinspection GoUnhandledErrorResult
		defer f.Close()
		if fd, err := f.Stat(); err != nil {
			instance.Status(ctx, fasthttp.StatusNotFound, p, false)
		} else if fd.IsDir() {
			instance.Status(ctx, fasthttp.StatusNotFound, p, false)
		} else {
			ext := path.Ext(p)
			contentType := mime.TypeByExtension(ext)
			if contentType == "" {
				contentType = "application/octet-stream"
			}
			ctx.Client.Response.Header().Set("Content-Type", contentType)
			ctx.Client.Status = http.StatusOK
			ctx.Result = context.ResultFallback
			ctx.Client.Response.WriteHeader(ctx.Client.Status)
			_, _ = io.Copy(ctx.Client.Response, f)
		}
	}
}
