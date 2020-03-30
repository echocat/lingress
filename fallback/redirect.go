package fallback

import (
	"fmt"
	"github.com/echocat/lingress/context"
	"github.com/echocat/lingress/support"
	"net/http"
	"strings"
)

func (instance *Fallback) Redirect(ctx *context.Context, statusCode int, target string) {
	if strings.ContainsAny(target, "\r\n") {
		p := ""
		if u, err := ctx.Client.RequestedUrl(); err == nil && u != nil {
			p = u.Path
		}
		instance.Status(ctx, http.StatusUnprocessableEntity, p, false)
		return
	}
	ctx.Client.Status = statusCode
	ctx.Client.Response.Header().Set("Location", support.NormalizeHeaderContent(target))
	if ctx.Client.Request.Method == "HEAD" {
		ctx.Client.Response.WriteHeader(ctx.Client.Status)
		return
	}

	lc := newLocationContextForCtx(ctx, instance.Bundle)
	switch support.NegotiateContentTypeOf(ctx.Client.Request, "application/x-yaml", "application/xml", "application/json", "text/html", "text/plain") {
	case "application/json":
		instance.RedirectJson(ctx, statusCode, target, lc)
	case "application/x-yaml":
		instance.RedirectYaml(ctx, statusCode, target, lc)
	case "application/xml":
		instance.RedirectXml(ctx, statusCode, target, lc)
	case "text/html":
		instance.RedirectHtml(ctx, statusCode, target, lc)
	default:
		instance.RedirectText(ctx, statusCode, target, lc)
	}
}

func (instance *Fallback) RedirectJson(ctx *context.Context, statusCode int, target string, lc *support.LocalizationContext) {
	genericResponseWithTarget(
		newLocalizedGenericResponse(ctx, statusCode, lc),
		target,
	).StreamJsonTo(ctx.Client.Response, ctx.Client.Request)
}

func (instance *Fallback) RedirectYaml(ctx *context.Context, statusCode int, target string, lc *support.LocalizationContext) {
	genericResponseWithTarget(
		newLocalizedGenericResponse(ctx, statusCode, lc),
		target,
	).StreamYamlTo(ctx.Client.Response, ctx.Client.Request)
}

func (instance *Fallback) RedirectXml(ctx *context.Context, statusCode int, target string, lc *support.LocalizationContext) {
	genericResponseWithTarget(
		newLocalizedGenericResponse(ctx, statusCode, lc),
		target,
	).StreamXmlTo(ctx.Client.Response, ctx.Client.Request)
}

func (instance *Fallback) RedirectText(ctx *context.Context, statusCode int, _ string, lc *support.LocalizationContext) {
	ctx.Client.Response.Header().Set("Content-Type", "text/plain; charset=utf-8")
	ctx.Client.Response.WriteHeader(statusCode)
	_, _ = ctx.Client.Response.Write([]byte(fmt.Sprintf("%d. %s\n", statusCode, localizeStatus(statusCode, lc))))
}

func (instance *Fallback) RedirectHtml(ctx *context.Context, statusCode int, target string, lc *support.LocalizationContext) {
	p := ""
	if u, err := ctx.Client.RequestedUrl(); err == nil && u != nil {
		p = u.Path
	}
	if tmpl, err := cloneAndLocalizeTemplate(instance.RedirectTemplate, lc); err != nil {
		return
	} else {
		object := map[string]interface{}{
			"path":       p,
			"statusCode": statusCode,
			"target":     target,
			"requestId":  ctx.Id.String(),
		}
		ctx.Client.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
		ctx.Client.Response.WriteHeader(statusCode)
		if err := tmpl.Execute(ctx.Client.Response, object); err != nil {
			support.LogForRequest(ctx.Client.Request).
				WithError(err).
				WithField("statusCode", statusCode).
				Error("could not render status page.")
		}
	}
}
