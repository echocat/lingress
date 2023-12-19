package settings

import (
	"github.com/echocat/lingress/support"
	"time"
)

func NewFallback() (Fallback, error) {
	return Fallback{
		ReloadTimeoutOnTemporaryIssues: 15 * time.Second,
	}, nil
}

type Fallback struct {
	ReloadTimeoutOnTemporaryIssues time.Duration `json:"reloadTimeoutOnTemporaryIssues,omitempty" yaml:"reloadTimeoutOnTemporaryIssues,omitempty"`
}

func (this *Fallback) RegisterFlags(fe support.FlagEnabled, appPrefix string) {
	fe.Flag("fallback.reloadTimeoutOnTemporaryIssues", "Timeout after which we try the reload the page on temporary issues.").
		Default(this.ReloadTimeoutOnTemporaryIssues.String()).
		Envar(support.FlagEnvName(appPrefix, "FALLBACK_RELOAD_TIMEOUT_ON_TEMPORARY_ISSUES")).
		DurationVar(&this.ReloadTimeoutOnTemporaryIssues)
}
