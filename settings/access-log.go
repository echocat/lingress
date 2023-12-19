package settings

import (
	"fmt"
	"github.com/echocat/lingress/support"
	"github.com/echocat/lingress/value"
)

func NewAccessLog() (AccessLog, error) {
	return AccessLog{
		QueueSize: 5000,
		Inline:    value.False(),
	}, nil
}

type AccessLog struct {
	QueueSize uint16     `yaml:"queueSize" json:"queueSize"`
	Inline    value.Bool `yaml:"inline" json:"inline"`
}

func (this *AccessLog) RegisterFlags(fe support.FlagEnabled, appPrefix string) {
	fe.Flag("accessLog.queueSize", "Maximum number of accessLog elements that could be queue before blocking.").
		PlaceHolder(fmt.Sprint(this.QueueSize)).
		Envar(support.FlagEnvName(appPrefix, "ACCESS_LOG_QUEUE_SIZE")).
		Uint16Var(&this.QueueSize)
	fe.Flag("accessLog.inline", "Instead of exploding the accessLog entries into sub-entries everything is inlined into the root object.").
		Envar(support.FlagEnvName(appPrefix, "ACCESS_LOG_INLINE")).
		SetValue(&this.Inline)
}
