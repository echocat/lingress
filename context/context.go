package context

import (
	"encoding/json"
	"github.com/echocat/lingress/rules"
	"github.com/echocat/lingress/server"
	"github.com/echocat/lingress/settings"
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
	Settings      *settings.Settings
	Client        Client
	Upstream      Upstream
	Id            Id
	CorrelationId Id
	Stage         Stage

	Logger log.Logger

	Rule   rules.Rule
	Result Result
	Error  error

	Properties map[string]interface{}
}

func AcquireContext(
	s *settings.Settings,
	connector server.ConnectorId,
	fromOtherReverseProxy bool,
	resp http.ResponseWriter,
	req *http.Request,
	logger log.Logger,
) (*Context, error) {
	success := false
	result := contextPool.Get().(*Context)
	defer func(created *Context) {
		if !success {
			created.Release()
		}
	}(result)

	result.Settings = s
	id, err := NewId(fromOtherReverseProxy, req)
	if err != nil {
		return nil, err
	}
	result.Id = id
	correlationId, err := NewCorrelationId(req)
	if err != nil {
		return nil, err
	}
	result.CorrelationId = correlationId
	result.Stage = StageCreated
	result.Client.configure(connector, fromOtherReverseProxy, resp, req)
	result.Upstream.configure()
	result.Logger = logger

	result.Rule = nil
	result.Result = ResultUnknown
	result.Error = nil

	result.Properties = make(map[string]interface{})

	success = true
	return result, nil
}

func (this *Context) Done(result Result, err ...error) {
	if len(err) > 1 {
		panic("there are only 0 or 1 errors allowed to be provided with this method")
	}
	if len(err) == 1 {
		this.Error = err[0]
	}
	this.Result = result
}

func (this *Context) MarkError(err error) {
	this.Error = err
	this.Client.Status = http.StatusInternalServerError
}

func (this *Context) MarkUnavailable(err error) {
	this.Error = err
	this.Client.Status = http.StatusServiceUnavailable
}

func (this *Context) MarkUnknown() {
	this.Client.Status = http.StatusNotFound
}

func (this *Context) Log() log.Logger {
	return this.Logger.WithAll(this.AsMap(false))
}

func (this *Context) LogProvider() func() log.Logger {
	return this.Log
}

func (this *Context) Release() {
	this.Client.clean()
	this.Upstream.clean()
	this.Stage = StageUnknown
	this.Id = NilRequestId
	this.CorrelationId = NilRequestId
	this.Logger = nil

	this.Rule = nil
	this.Result = ResultUnknown
	this.Error = nil

	this.Properties = nil

	contextPool.Put(this)
}

func (this *Context) MarshalJSON() ([]byte, error) {
	return json.Marshal(this.AsMap(false))
}

func (this *Context) AsMap(inlineFields bool) map[string]interface{} {
	const (
		clientSubPrefix   = FieldClient + "."
		upstreamSubPrefix = FieldUpstream + "."
	)
	buf := map[string]interface{}{
		FieldRequestId:     this.Id,
		FieldCorrelationId: this.CorrelationId,
		FieldResult:        this.Result.Name(),
	}
	if inlineFields {
		this.Client.ApplyToMap(clientSubPrefix, &buf)
		this.Upstream.ApplyToMap(this.Rule, upstreamSubPrefix, &buf)
	} else {
		buf[FieldClient] = this.Client.AsMap()
		if b := this.Upstream.AsMap(this.Rule); len(b) > 0 {
			buf[FieldUpstream] = b
		}
	}
	if err := this.Error; err != nil {
		buf[FieldError] = err
	}
	return buf
}
