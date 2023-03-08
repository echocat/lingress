package support

import (
	_ "github.com/echocat/slf4g-klog/bridge/hook"
	"github.com/echocat/slf4g/native"
	"github.com/echocat/slf4g/native/facade/value"
)

var (
	loggerValue = value.NewProvider(native.DefaultProvider)
	_           = RegisterFlagRegistrar(&logFacade{})
)

type logFacade struct{}

func (instance *logFacade) RegisterFlag(fe FlagEnabled, appPrefix string) error {
	Must(loggerValue.Consumer.Formatter.Set("json"))

	fe.Flag("logLevel", "").
		PlaceHolder("<log level; default: " + loggerValue.Level.String() + ">").
		Envar(FlagEnvName(appPrefix, "LOG_LEVEL")).
		SetValue(loggerValue.Level)
	fe.Flag("logFormat", "").
		PlaceHolder("<log format; default: " + loggerValue.Level.String() + ">").
		Envar(FlagEnvName(appPrefix, "LOG_FORMAT")).
		SetValue(loggerValue.Consumer.Formatter)
	fe.Flag("logColor", "").
		PlaceHolder("<log color; default: " + loggerValue.Consumer.Formatter.ColorMode.String() + ">").
		Envar(FlagEnvName(appPrefix, "LOG_COLOR")).
		SetValue(loggerValue.Consumer.Formatter.ColorMode)

	return nil
}
