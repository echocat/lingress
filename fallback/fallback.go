package fallback

import (
	"fmt"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/echocat/lingress/support"
	"html/template"
	"regexp"
	"time"
)

type Fallback struct {
	redirectTemplate *template.Template
	statusTemplate   *template.Template
	bundle           *i18n.Bundle

	fixTarget                string
	fixTargetForHostsPattern *regexp.Regexp
	fixTargetToHttps         bool

	pathPrefix                     string
	reloadTimeoutOnTemporaryIssues time.Duration
}

func New() (*Fallback, error) {
	result := Fallback{
		fixTarget:                "/view/",
		fixTargetForHostsPattern: nil,
		fixTargetToHttps:         false,

		pathPrefix:                     "",
		reloadTimeoutOnTemporaryIssues: time.Second * 15,
	}

	sTmpl, err := newTemplate("status.html", template.FuncMap{
		"isStatusTemporaryIssue":  isStatusTemporaryIssue,
		"isStatusCodeAnIssue":     isStatusCodeAnIssue,
		"isStatusClientSideIssue": isStatusClientSideIssue,
		"isStatusServerSideIssue": isStatusServerSideIssue,
	})
	if err != nil {
		return nil, err
	}
	result.statusTemplate = sTmpl

	rTmpl, err := newTemplate("redirect.html")
	if err != nil {
		return nil, err
	}
	result.redirectTemplate = rTmpl

	bundle, err := newBundle()
	if err != nil {
		return nil, err
	}
	result.bundle = bundle

	return &result, nil
}

func (instance *Fallback) RegisterFlag(fe support.FlagEnabled, appPrefix string) error {
	fe.Flag("fixTarget", "Target where to redirect to if someone accesses / or other stuff.").
		PlaceHolder(instance.fixTarget).
		Envar(support.FlagEnvName(appPrefix, "FIX_TARGET")).
		StringVar(&instance.fixTarget)
	fe.Flag("fixTargetForHostPattern", "Only do this target fixing for hosts that matches this pattern.").
		PlaceHolder("<regexp>").
		Envar(support.FlagEnvName(appPrefix, "FIX_TARGET_FOR_HOST_PATTERN")).
		RegexpVar(&instance.fixTargetForHostsPattern)
	fe.Flag("fixTargetToHttps", "If enabled and a target will be fixed it will be enforced to be HTTPS.").
		PlaceHolder(fmt.Sprint(instance.fixTargetToHttps)).
		Envar(support.FlagEnvName(appPrefix, "FIX_TARGET_TO_HTTPS")).
		BoolVar(&instance.fixTargetToHttps)

	fe.Flag("pathPrefix", "Appends the rendered paths with this prefix.").
		Default(instance.pathPrefix).
		Envar(support.FlagEnvName(appPrefix, "PATH_PREFIX")).
		StringVar(&instance.pathPrefix)
	fe.Flag("reloadTimeoutOnTemporaryIssues", "Timeout after which we try the reload the page on temporary issues.").
		Default(instance.reloadTimeoutOnTemporaryIssues.String()).
		Envar(support.FlagEnvName(appPrefix, "RELOAD_TIMEOUT_ON_TEMPORARY_ISSUES")).
		DurationVar(&instance.reloadTimeoutOnTemporaryIssues)

	return nil
}
