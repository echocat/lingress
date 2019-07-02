package fallback

import (
	"github.com/gobuffalo/packr"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/echocat/lingress/context"
	"github.com/echocat/lingress/support"
)

var (
	localizations = packr.NewBox("../resources/localization")
)

func newBundle() (bundle *i18n.Bundle, err error) {
	return support.LoadBundle(localizations)
}

func newLocationContextForCtx(ctx *context.Context, bundle *i18n.Bundle) *support.LocalizationContext {
	return &support.LocalizationContext{
		Bundle:         bundle,
		AcceptLanguage: string(ctx.Client.Request.Header.Get("Accept-Language")),
	}
}

func newLocalizedGenericResponse(ctx *context.Context, statusCode int, lc *support.LocalizationContext) support.GenericResponse {
	message := localizeStatus(statusCode, lc)
	return support.NewGenericResponse(statusCode, message, ctx.Client.Request)
}

func genericResponseWithTarget(in support.GenericResponse, target string) support.GenericResponse {
	return in.WithData(struct {
		Target string `json:"target" yaml:"target" xml:"target"`
	}{
		Target: target,
	})
}
