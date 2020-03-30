package fallback

import (
	"github.com/echocat/lingress/support"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"html/template"
	"time"
)

type Fallback struct {
	FileProviders    support.FileProviders
	RedirectTemplate *template.Template
	StatusTemplate   *template.Template
	Bundle           *i18n.Bundle

	ReloadTimeoutOnTemporaryIssues time.Duration
}

func New(fps support.FileProviders) (*Fallback, error) {
	result := Fallback{
		FileProviders:                  fps,
		ReloadTimeoutOnTemporaryIssues: time.Second * 15,
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

func (instance *Fallback) RegisterFlag(fe support.FlagEnabled, appPrefix string) error {
	fe.Flag("reloadTimeoutOnTemporaryIssues", "Timeout after which we try the reload the page on temporary issues.").
		Default(instance.ReloadTimeoutOnTemporaryIssues.String()).
		Envar(support.FlagEnvName(appPrefix, "RELOAD_TIMEOUT_ON_TEMPORARY_ISSUES")).
		DurationVar(&instance.ReloadTimeoutOnTemporaryIssues)

	return nil
}
