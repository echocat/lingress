package fallback

import (
	"github.com/echocat/lingress/file/providers"
	"github.com/echocat/lingress/settings"
	log "github.com/echocat/slf4g"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"html/template"
)

type Fallback struct {
	settings *settings.Settings

	FileProviders    providers.FileProviders
	RedirectTemplate *template.Template
	StatusTemplate   *template.Template
	Bundle           *i18n.Bundle
	Logger           log.Logger
}

func New(s *settings.Settings, fps providers.FileProviders, logger log.Logger) (*Fallback, error) {
	result := Fallback{
		settings:      s,
		FileProviders: fps,
		Logger:        logger,
	}

	sTmpl, err := newTemplate(fps.GetTemplates(), "status.html", template.FuncMap{
		"isStatusTemporaryIssue":  isStatusTemporaryIssue,
		"isStatusCodeAnIssue":     isStatusCodeAnIssue,
		"isStatusClientSideIssue": isStatusClientSideIssue,
		"isStatusServerSideIssue": isStatusServerSideIssue,
	})
	if err != nil {
		return nil, err
	}
	result.StatusTemplate = sTmpl

	rTmpl, err := newTemplate(fps.GetTemplates(), "redirect.html")
	if err != nil {
		return nil, err
	}
	result.RedirectTemplate = rTmpl

	bundle, err := newBundle(fps.GetLocalization())
	if err != nil {
		return nil, err
	}
	result.Bundle = bundle

	return &result, nil
}
