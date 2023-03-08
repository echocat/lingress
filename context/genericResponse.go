package context

import (
	"github.com/echocat/lingress/support"
	"net/http"
	"time"
)

type GenericResponse struct {
	context *Context
	value   support.GenericResponse
}

func (instance *Context) NewGenericResponse(statusCode int, message string) *GenericResponse {
	var path string
	if u, err := instance.Client.RequestedUrl(); err != nil || u == nil {
		// ignore
	} else {
		path = u.Path
	}
	result := &GenericResponse{
		context: instance,
		value: support.GenericResponse{
			Timestamp:     time.Now().Format(time.RFC3339),
			Status:        statusCode,
			Message:       message,
			Path:          path,
			RequestId:     instance.Id.String(),
			CorrelationId: instance.CorrelationId.String(),
		},
	}
	result.value.ErrorHandler = result.errorHandler
	return result
}

func (instance *GenericResponse) errorHandler(_ http.ResponseWriter, _ *http.Request, message string, err error, status int) {
	instance.context.Log().
		WithError(err).
		With("statusCode", status).
		Error(message)
}

func (instance *GenericResponse) SetPath(path string) *GenericResponse {
	instance.value.Path = path
	return instance
}

func (instance *GenericResponse) SetData(data interface{}) *GenericResponse {
	instance.value.Data = data
	return instance
}

func (instance *GenericResponse) StreamAsJson() {
	instance.value.StreamJsonTo(instance.context.Client.Response, instance.context.Client.Request)
}

func (instance *GenericResponse) StreamAsYaml() {
	instance.value.StreamYamlTo(instance.context.Client.Response, instance.context.Client.Request)
}

func (instance GenericResponse) StreamAsXml() {
	instance.value.StreamXmlTo(instance.context.Client.Response, instance.context.Client.Request)
}
