package slf4g_native

import (
	"github.com/echocat/slf4g"
	"github.com/echocat/slf4g-klog/bridge"
	_ "github.com/echocat/slf4g-klog/bridge/hook"
	"k8s.io/apimachinery/pkg/util/runtime"
)

func init() {
	kl := log.GetLogger("kubernetes")
	bridge.ConfigureWith(kl)
	runtime.ErrorHandlers = []func(err error){func(err error) {
		kl.WithError(err).Error("Unhandled kubernetes internal error.")
	}}
}
