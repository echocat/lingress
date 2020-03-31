package context

import (
	"encoding/json"
	"github.com/echocat/lingress/rules"
	"github.com/echocat/lingress/support"
	log "github.com/sirupsen/logrus"
	"net/http"
	"sync"
	"time"
)

var (
	contextPool = sync.Pool{
		New: func() interface{} {
			return new(Context)
		},
	}
	emptyTime = time.Time{}
)

type Context struct {
	Client   Client
	Upstream Upstream
	Id       Id
	Stage    Stage

	Rule   rules.Rule
	Result Result
	Error  error
}

func AcquireContext(fromOtherReverseProxy bool, resp http.ResponseWriter, req *http.Request) *Context {
	success := false
	result := contextPool.Get().(*Context)
	defer func(created *Context) {
		if !success {
			created.Release()
		}
	}(result)

	result.Id = NewId(req)
	result.Stage = StageCreated
	result.Client.configure(fromOtherReverseProxy, resp, req)
	result.Upstream.configure()

	result.Rule = nil
	result.Result = ResultUnknown
	result.Error = nil

	success = true
	return result
}

func (instance *Context) Done(result Result, err ...error) {
	if len(err) > 1 {
		panic("there are only 0 or 1 errors allowed to be provided with this method")
	}
	if len(err) == 1 {
		instance.Error = err[0]
	}
	instance.Result = result
}

func (instance *Context) MarkError(err error) {
	instance.Error = err
	instance.Client.Status = http.StatusInternalServerError
}

func (instance *Context) MarkUnavailable(err error) {
	instance.Error = err
	instance.Client.Status = http.StatusServiceUnavailable
}

func (instance *Context) MarkUnknown() {
	instance.Client.Status = http.StatusNotFound
}

func (instance *Context) Log() log.FieldLogger {
	return log.WithFields(instance.AsMap())
}

func (instance *Context) Release() {
	instance.Client.clean()
	instance.Upstream.clean()
	instance.Stage = StageUnknown
	instance.Id = NilRequestId

	instance.Rule = nil
	instance.Result = ResultUnknown
	instance.Error = nil

	contextPool.Put(instance)
}

func (instance *Context) MarshalJSON() ([]byte, error) {
	return json.Marshal(instance.AsMap())
}

func (instance *Context) AsMap() map[string]interface{} {
	buf := map[string]interface{}{
		"id":      instance.Id,
		"client":  instance.Client.AsMap(),
		"runtime": support.Runtime(),
		"result":  instance.Result.Name(),
	}
	if err := instance.Error; err != nil {
		buf["error"] = err
	}
	if b := instance.Upstream.AsMap(instance.Rule); len(b) > 0 {
		buf["upstream"] = b
	}
	return buf
}
