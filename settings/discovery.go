package settings

import (
	"github.com/echocat/lingress/support"
	"time"
)

func NewDiscovery() (Discovery, error) {
	return Discovery{
		ResyncAfter: 10 * time.Minute,
	}, nil
}

type Discovery struct {
	ResyncAfter time.Duration `yaml:"resyncAfter,omitempty" json:"resyncAfter,omitempty"`
}

func (this *Discovery) RegisterFlags(fe support.FlagEnabled, appPrefix string) {
	fe.Flag("discovery.resyncAfter", "Time after which the configuration should be resynced to be ensure to be not out of date.").
		PlaceHolder(this.ResyncAfter.String()).
		Envar(support.FlagEnvName(appPrefix, "DISCOVERY_RESYNC_AFTER")).
		DurationVar(&this.ResyncAfter)
}
