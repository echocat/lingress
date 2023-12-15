package fallback

import (
	"fmt"
	"github.com/echocat/lingress/context"
	"github.com/echocat/lingress/i18n"
	"github.com/echocat/lingress/support"
	"net/http"
	"strings"
)

func (this *Fallback) Redirect(ctx *context.Context, statusCode int, target string) {
	if strings.ContainsAny(target, "\r\n") {
		p := ""
		if u, err := ctx.Client.RequestedUrl(); err == nil && u != nil {
			p = u.Path
		}
		this.Status(ctx, http.StatusUnprocessableEntity, p, false)
		return
	}
	ctx.Client.Status = statusCode
	ctx.Client.Response.Header().Set("Location", support.NormalizeHeaderContent(target))
	if ctx.Client.Request.Method == "HEAD" {
		ctx.Client.Response.WriteHeader(ctx.Client.Status)
		return
	}

	lc := newLocationContextForCtx(ctx, this.Bundle)
	switch support.NegotiateContentTypeOf(ctx.Client.Request, "application/x-yaml", "application/xml", "application/json", "text/html", "text/plain") {
	case "application/json":
		this.RedirectJson(ctx, statusCode, target, lc)
	case "application/x-yaml":
		this.RedirectYaml(ctx, statusCode, target, lc)
	case "application/xml":
		this.RedirectXml(ctx, statusCode, target, lc)
	case "text/html":
		this.RedirectHtml(ctx, statusCode, target, lc)
	default:
		this.RedirectText(ctx, statusCode, target, lc)
	}
}

func (this *Fallback) RedirectJson(ctx *context.Context, statusCode int, target string, lc *i18n.LocalizationContext) {
	genericResponseWithTarget(
		newLocalizedGenericResponse(ctx, statusCode, lc),
		target,
	).StreamAsJson()
}

func (this *Fallback) RedirectYaml(ctx *context.Context, statusCode int, target string, lc *i18n.LocalizationContext) {
	genericResponseWithTarget(
		newLocalizedGenericResponse(ctx, statusCode, lc),
		target,
	).StreamAsYaml()
}

func (this *Fallback) RedirectXml(ctx *context.Context, statusCode int, target string, lc *i18n.LocalizationContext) {
	genericResponseWithTarget(
		newLocalizedGenericResponse(ctx, statusCode, lc),
		target,
	).StreamAsXml()
}

func (this *Fallback) RedirectText(ctx *context.Context, statusCode int, _ string, lc *i18n.LocalizationContext) {
	ctx.Client.Response.Header().Set("Content-Type", "text/plain; charset=utf-8")
	ctx.Client.Response.WriteHeader(statusCode)
	_, _ = ctx.Client.Response.Write([]byte(fmt.Sprintf("%d. %s\n", statusCode, localizeStatus(statusCode, lc))))
}

func (this *Fallback) RedirectHtml(ctx *context.Context, statusCode int, target string, lc *i18n.LocalizationContext) {
	p := ""
	if u, err := ctx.Client.RequestedUrl(); err == nil && u != nil {
		p = u.Path
	}
	if tmpl, err := cloneAndLocalizeTemplate(this.RedirectTemplate, lc); err != nil {
		return
	} else {
		object := map[string]interface{}{
			"path":          p,
			"statusCode":    statusCode,
			"target":        target,
			"requestId":     ctx.Id.String(),
			"correlationId": ctx.CorrelationId.String(),
		}
		ctx.Client.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
		ctx.Client.Response.WriteHeader(statusCode)
		if err := tmpl.Execute(ctx.Client.Response, object); err != nil {
			ctx.Log().
				WithError(err).
				With("statusCode", statusCode).
				Error("Could not render status page.")
		}
	}
}
