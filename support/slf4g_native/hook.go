package slf4g_native

import (
	"context"
	"fmt"
	"github.com/echocat/slf4g"
	"github.com/echocat/slf4g-klog/bridge"
	_ "github.com/echocat/slf4g-klog/bridge/hook"
	"k8s.io/apimachinery/pkg/util/runtime"
)

func init() {
	kl := log.GetLogger("kubernetes")
	bridge.ConfigureWith(kl)
	runtime.ErrorHandlers = []runtime.ErrorHandler{func(_ context.Context, err error, msg string, keysAndValues ...interface{}) {
		out := "Unhandled kubernetes internal error"
		if msg != "" {
			out += ": " + fmt.Sprintf(msg, keysAndValues...)
		} else {
			out += "."
		}
		kl.WithError(err).Error(out)
	}}
}
