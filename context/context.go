package context

import (
	"encoding/json"
	"github.com/echocat/lingress/rules"
	"github.com/echocat/lingress/server"
	"github.com/echocat/lingress/support"
	"github.com/echocat/slf4g"
	"net/http"
	"sync"
	"time"
)

const (
	FieldRequestId     = "requestId"
	FieldCorrelationId = "correlationId"
	FieldClient        = "client"
	FieldUpstream      = "upstream"
	FieldRuntime       = "runtime"
	FieldResult        = "result"
	FieldError         = "error"
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
	Client        Client
	Upstream      Upstream
	Id            Id
	CorrelationId Id
	Stage         Stage

	Rule   rules.Rule
	Result Result
	Error  error

	Properties map[string]interface{}
}

func AcquireContext(connector server.ConnectorId, fromOtherReverseProxy bool, resp http.ResponseWriter, req *http.Request) *Context {
	success := false
	result := contextPool.Get().(*Context)
	defer func(created *Context) {
		if !success {
			created.Release()
		}
	}(result)

	result.Id = NewId(fromOtherReverseProxy, req)
	result.CorrelationId = NewCorrelationId(fromOtherReverseProxy, req)
	result.Stage = StageCreated
	result.Client.configure(connector, fromOtherReverseProxy, resp, req)
	result.Upstream.configure()

	result.Rule = nil
	result.Result = ResultUnknown
	result.Error = nil

	result.Properties = make(map[string]interface{})

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

func (instance *Context) Log() log.Logger {
	return log.WithAll(instance.AsMap(false))
}

func (instance *Context) Release() {
	instance.Client.clean()
	instance.Upstream.clean()
	instance.Stage = StageUnknown
	instance.Id = NilRequestId
	instance.CorrelationId = NilRequestId

	instance.Rule = nil
	instance.Result = ResultUnknown
	instance.Error = nil

	instance.Properties = nil

	contextPool.Put(instance)
}

func (instance *Context) MarshalJSON() ([]byte, error) {
	return json.Marshal(instance.AsMap(false))
}

func (instance *Context) AsMap(inlineFields bool) map[string]interface{} {
	const (
		clientSubPrefix   = FieldClient + "."
		upstreamSubPrefix = FieldUpstream + "."
	)
	buf := map[string]interface{}{
		FieldRequestId:     instance.Id,
		FieldCorrelationId: instance.CorrelationId,
		FieldRuntime:       support.Runtime(),
		FieldResult:        instance.Result.Name(),
	}
	if inlineFields {
		instance.Client.ApplyToMap(clientSubPrefix, &buf)
		instance.Upstream.ApplyToMap(instance.Rule, upstreamSubPrefix, &buf)
	} else {
		buf[FieldClient] = instance.Client.AsMap()
		if b := instance.Upstream.AsMap(instance.Rule); len(b) > 0 {
			buf[FieldUpstream] = b
		}
	}
	if err := instance.Error; err != nil {
		buf[FieldError] = err
	}
	return buf
}
