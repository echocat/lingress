package fallback

import (
	"fmt"
	"github.com/echocat/lingress/context"
	"github.com/echocat/lingress/i18n"
	"github.com/echocat/lingress/support"
	"math"
	"time"
)

func (this *Fallback) Status(ctx *context.Context, statusCode int, path string, canHandleTemporary bool) {
	ctx.Client.Status = statusCode
	if ctx.Client.Request.Method == "HEAD" {
		ctx.Client.Response.WriteHeader(statusCode)
		return
	}
	lc := newLocationContextForCtx(ctx, this.Bundle)
	switch support.NegotiateContentTypeOf(ctx.Client.Request, "application/x-yaml", "application/xml", "application/json", "text/html", "text/plain") {
	case "application/json":
		this.StatusAsJson(ctx, statusCode, path, canHandleTemporary, lc)
	case "application/x-yaml":
		this.StatusAsYaml(ctx, statusCode, path, canHandleTemporary, lc)
	case "application/xml":
		this.StatusAsXml(ctx, statusCode, path, canHandleTemporary, lc)
	case "text/html":
		this.StatusAsHtml(ctx, statusCode, path, canHandleTemporary, lc)
	default:
		this.StatusAsText(ctx, statusCode, path, canHandleTemporary, lc)
	}
}

func (this *Fallback) StatusAsJson(ctx *context.Context, statusCode int, path string, _ bool, lc *i18n.LocalizationContext) {
	newLocalizedGenericResponse(ctx, statusCode, lc).
		SetPath(path).
		StreamAsJson()
}

func (this *Fallback) StatusAsYaml(ctx *context.Context, statusCode int, path string, _ bool, lc *i18n.LocalizationContext) {
	newLocalizedGenericResponse(ctx, statusCode, lc).
		SetPath(path).
		StreamAsYaml()
}

func (this *Fallback) StatusAsXml(ctx *context.Context, statusCode int, path string, _ bool, lc *i18n.LocalizationContext) {
	newLocalizedGenericResponse(ctx, statusCode, lc).
		SetPath(path).
		StreamAsXml()
}

func (this *Fallback) StatusAsText(ctx *context.Context, statusCode int, _ string, _ bool, lc *i18n.LocalizationContext) {
	ctx.Client.Response.Header().Set("Content-Type", "text/plain; charset=utf-8")
	ctx.Client.Response.WriteHeader(statusCode)
	_, _ = ctx.Client.Response.Write([]byte(fmt.Sprintf("%d. %s\n", statusCode, localizeStatus(statusCode, lc))))
}

func (this *Fallback) StatusAsHtml(ctx *context.Context, statusCode int, path string, canHandleTemporary bool, lc *i18n.LocalizationContext) {
	if tmpl, err := cloneAndLocalizeTemplate(this.StatusTemplate, lc); err != nil {
		ctx.Client.Response.WriteHeader(statusCode)
		return
	} else {
		object := map[string]interface{}{
			"statusCode":         statusCode,
			"path":               path,
			"autoReloadSeconds":  int(math.Round(this.settings.Fallback.ReloadTimeoutOnTemporaryIssues.Seconds())),
			"canHandleTemporary": canHandleTemporary,
			"year":               time.Now().Year(),
			"requestId":          ctx.Id.String(),
			"correlationId":      ctx.CorrelationId.String(),
		}
		ctx.Client.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
		ctx.Client.Response.WriteHeader(statusCode)
		if err := tmpl.Execute(ctx.Client.Response, object); err != nil {
			ctx.Log().
				WithError(err).
				With("statusCode", statusCode).
				Error("could not render status page.")
		}
	}
}
