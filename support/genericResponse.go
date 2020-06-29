package support

import (
	"encoding/json"
	"encoding/xml"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"net/http"
	"time"
)

type GenericResponse struct {
	XMLName       xml.Name    `json:"-" yaml:"-" xml:"response"`
	Timestamp     string      `json:"timestamp" yaml:"timestamp" xml:"timestamp"`
	Status        int         `json:"status" yaml:"status" xml:"status"`
	Message       string      `json:"message,omitempty" yaml:"message,omitempty" xml:"message,omitempty"`
	Path          string      `json:"path" yaml:"path" xml:"path"`
	RequestId     string      `json:"requestId,omitempty" yaml:"requestId,omitempty" xml:"requestId,omitempty"`
	CorrelationId string      `json:"correlationId,omitempty" yaml:"correlationId,omitempty" xml:"correlationId,omitempty"`
	Data          interface{} `json:"data,omitempty" yaml:"data,omitempty" xml:"data,omitempty"`

	ErrorHandler func(resp http.ResponseWriter, req *http.Request, message string, err error, status int)
}

func NewGenericResponse(statusCode int, message string, req *http.Request) GenericResponse {
	return GenericResponse{
		Timestamp: time.Now().Format(time.RFC3339),
		Status:    statusCode,
		Message:   message,
		Path:      PathOfRequest(req),
		RequestId: RequestIdOfRequest(req),
	}
}

func (instance GenericResponse) WithData(data interface{}) GenericResponse {
	result := instance
	result.Data = data
	return result
}

func (instance GenericResponse) StreamJsonTo(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(instance.Status)
	j := json.NewEncoder(resp)
	j.SetIndent("", "  ")
	if err := j.Encode(instance.Data); err != nil {
		instance.logError(resp, req, "Could not render json response.", err)
	}
}

func (instance GenericResponse) StreamYamlTo(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Set("Content-Type", "application/x-yaml")
	resp.WriteHeader(instance.Status)
	y := yaml.NewEncoder(resp)
	if err := y.Encode(instance.Data); err != nil {
		instance.logError(resp, req, "Could not render yml response.", err)
	}
}

func (instance GenericResponse) StreamXmlTo(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Set("Content-Type", "application/xml")
	resp.WriteHeader(instance.Status)
	x := xml.NewEncoder(resp)
	x.Indent("", "  ")
	if err := x.Encode(instance.Data); err != nil {
		instance.logError(resp, req, "Could not render xml response.", err)
	}
}

func (instance GenericResponse) logError(resp http.ResponseWriter, req *http.Request, message string, err error) {
	if f := instance.ErrorHandler; f != nil {
		f(resp, req, message, err, instance.Status)
		return
	}
	log.WithFields(log.Fields{
		"runtime":       Runtime(),
		"requestId":     instance.RequestId,
		"correlationId": instance.CorrelationId,
		"remoteIp":      RequestBasedLazyStringerFor(req, RemoteIpOfRequest),
		"host":          RequestBasedLazyStringerFor(req, HostOfRequest),
		"method":        req.Method,
		"requestUri":    RequestBasedLazyStringerFor(req, UriOfRequest),
		"userAgent":     RequestBasedLazyStringerFor(req, UserAgentOfRequest),
	}).WithError(err).Error(message)
}
