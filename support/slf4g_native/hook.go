package slf4g_native

import (
	"github.com/echocat/lingress/support"
	"github.com/echocat/slf4g"
	"github.com/echocat/slf4g-klog/bridge"
	_ "github.com/echocat/slf4g-klog/bridge/hook"
	"github.com/echocat/slf4g/native"
	"github.com/echocat/slf4g/native/facade/value"
	"k8s.io/apimachinery/pkg/util/runtime"
)

var hookVal = &hook{
	value.NewProvider(native.DefaultProvider),
}

func init() {
	support.RegisterFlagRegistrar(hookVal)

	kl := log.GetLogger("kubernetes")
	bridge.ConfigureWith(kl)
	runtime.ErrorHandlers = []func(err error){func(err error) {
		kl.WithError(err).Error("unhandled kubernetes internal error.")
	}}

}

type hook struct {
	value.Provider
}

func (instance *hook) RegisterFlag(fe support.FlagEnabled, appPrefix string) error {
	fe.Flag("log.level", "").
		Envar(support.FlagEnvName(appPrefix, "LOG_LEVEL")).
		SetValue(&instance.Provider.Level)
	fe.Flag("log.format", "").
		Default("text").
		Envar(support.FlagEnvName(appPrefix, "LOG_FORMAT")).
		SetValue(&instance.Provider.Consumer.Formatter)
	fe.Flag("log.color", "").
		Default("auto").
		Envar(support.FlagEnvName(appPrefix, "LOG_COLOR")).
		SetValue(instance.Provider.Consumer.Formatter.ColorMode)
	return nil
}
