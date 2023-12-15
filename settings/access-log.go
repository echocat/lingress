package settings

import (
	"fmt"
	"github.com/echocat/lingress/support"
)

func NewAccessLog() (AccessLog, error) {
	return AccessLog{
		QueueSize: 5000,
		Inline:    false,
	}, nil
}

type AccessLog struct {
	QueueSize uint16 `yaml:"queueSize" json:"queueSize"`
	Inline    bool   `yaml:"inline" json:"inline"`
}

func (this *AccessLog) RegisterFlags(fe support.FlagEnabled, appPrefix string) {
	fe.Flag("accessLog.queueSize", "Maximum number of accessLog elements that could be queue before blocking.").
		PlaceHolder(fmt.Sprint(this.QueueSize)).
		Envar(support.FlagEnvName(appPrefix, "ACCESS_LOG_QUEUE_SIZE")).
		Uint16Var(&this.QueueSize)
	fe.Flag("accessLog.inline", "Instead of exploding the accessLog entries into sub-entries everything is inlined into the root object.").
		Envar(support.FlagEnvName(appPrefix, "ACCESS_LOG_INLINE")).
		BoolVar(&this.Inline)
}
