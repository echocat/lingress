package support

import (
	"github.com/echocat/slf4g/native"
	"github.com/echocat/slf4g/native/formatter"
)

var (
	_ = RegisterFlagRegistrar(&logLevelFacade{})
)

type logLevelFacade struct{}

func (instance *logLevelFacade) RegisterFlag(fe FlagEnabled, appPrefix string) error {
	fe.Flag("logLevel", "On which level the output should be logged").
		PlaceHolder("<log level; default: " + native.DefaultProvider.Level.String() + ">").
		Envar(FlagEnvName(appPrefix, "LOG_LEVEL")).
		SetValue(native.DefaultProvider.Level)
	return nil
}

func init() {
	native.DefaultProvider.EventFormatter = formatter.DefaultJson
}
