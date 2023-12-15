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

func (this *Context) NewGenericResponse(statusCode int, message string) *GenericResponse {
	var path string
	if u, err := this.Client.RequestedUrl(); err != nil || u == nil {
		// ignore
	} else {
		path = u.Path
	}
	result := &GenericResponse{
		context: this,
		value: support.GenericResponse{
			Timestamp:     time.Now().Format(time.RFC3339),
			Status:        statusCode,
			Message:       message,
			Path:          path,
			RequestId:     this.Id.String(),
			CorrelationId: this.CorrelationId.String(),
		},
	}
	result.value.ErrorHandler = result.errorHandler
	return result
}

func (this *GenericResponse) errorHandler(_ http.ResponseWriter, _ *http.Request, message string, err error, status int) {
	this.context.Log().
		WithError(err).
		With("statusCode", status).
		Error(message)
}

func (this *GenericResponse) SetPath(path string) *GenericResponse {
	this.value.Path = path
	return this
}

func (this *GenericResponse) SetData(data interface{}) *GenericResponse {
	this.value.Data = data
	return this
}

func (this *GenericResponse) StreamAsJson() {
	this.value.StreamJsonTo(this.context.Client.Response, this.context.Client.Request, this.context.Log)
}

func (this *GenericResponse) StreamAsYaml() {
	this.value.StreamYamlTo(this.context.Client.Response, this.context.Client.Request, this.context.Log)
}

func (this GenericResponse) StreamAsXml() {
	this.value.StreamXmlTo(this.context.Client.Response, this.context.Client.Request, this.context.Log)
}
