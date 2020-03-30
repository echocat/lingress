package fallback

import (
	"github.com/echocat/lingress/context"
	"io"
	"mime"
	"net/http"
	"path"
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

	instance.fallback(ctx)
}

func (instance *Fallback) fallback(ctx *context.Context) {
	if p := ctx.Client.Request.RequestURI; !path.IsAbs(p) || path.Clean(p) != p {
		instance.Status(ctx, http.StatusNotFound, p, false)
	} else if fp, err := instance.FileProviders.GetStatic().Open(p); err != nil {
		instance.Status(ctx, http.StatusNotFound, p, false)
	} else {
		//noinspection GoUnhandledErrorResult
		defer fp.Close()
		if fi, err := fp.Stat(); err != nil || fi.IsDir() {
			instance.Status(ctx, http.StatusNotFound, p, false)
			return
		}
		ext := path.Ext(p)
		contentType := mime.TypeByExtension(ext)
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		ctx.Client.Response.Header().Set("Content-Type", contentType)
		ctx.Client.Status = http.StatusOK
		ctx.Result = context.ResultFallback
		ctx.Client.Response.WriteHeader(ctx.Client.Status)
		_, _ = io.Copy(ctx.Client.Response, fp)
	}
}
