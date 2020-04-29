package fallback

import (
	"github.com/echocat/lingress/context"
	"github.com/echocat/lingress/support"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

func newBundle(fp support.FileProvider) (bundle *i18n.Bundle, err error) {
	return support.LoadBundle(fp)
}

func newLocationContextForCtx(ctx *context.Context, bundle *i18n.Bundle) *support.LocalizationContext {
	return &support.LocalizationContext{
		Bundle:         bundle,
		AcceptLanguage: ctx.Client.Request.Header.Get("Accept-Language"),
	}
}

func newLocalizedGenericResponse(ctx *context.Context, statusCode int, lc *support.LocalizationContext) *context.GenericResponse {
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
