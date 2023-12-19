package fallback

import (
	"github.com/echocat/lingress/context"
	"github.com/echocat/lingress/file/providers"
	i18n2 "github.com/echocat/lingress/i18n"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

func newBundle(fp providers.FileProvider) (bundle *i18n.Bundle, err error) {
	return i18n2.LoadBundle(fp)
}

func newLocationContextForCtx(ctx *context.Context, bundle *i18n.Bundle) *i18n2.LocalizationContext {
	return &i18n2.LocalizationContext{
		Bundle:         bundle,
		AcceptLanguage: ctx.Client.Request.Header.Get("Accept-Language"),
		Logger:         ctx.Logger,
	}
}

func newLocalizedGenericResponse(ctx *context.Context, statusCode int, lc *i18n2.LocalizationContext) *context.GenericResponse {
	message := localizeStatus(statusCode, lc)
	return ctx.NewGenericResponse(statusCode, message)
}

func genericResponseWithTarget(in *context.GenericResponse, target string) *context.GenericResponse {
	return in.SetData(struct {
		Target string `json:"target" yaml:"target" xml:"target"`
	}{
		Target: target,
	})
}
