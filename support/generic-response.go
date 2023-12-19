package support

import (
	"encoding/json"
	"encoding/xml"
	"github.com/echocat/slf4g"
	"gopkg.in/yaml.v3"
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

func (this GenericResponse) WithData(data interface{}) GenericResponse {
	result := this
	result.Data = data
	return result
}

func (this GenericResponse) StreamJsonTo(resp http.ResponseWriter, req *http.Request, logProvider func() log.Logger) {
	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(this.Status)
	j := json.NewEncoder(resp)
	j.SetIndent("", "  ")
	if err := j.Encode(this.Data); err != nil {
		this.logError(resp, req, "Could not render json response.", err, logProvider)
	}
}

func (this GenericResponse) StreamYamlTo(resp http.ResponseWriter, req *http.Request, logProvider func() log.Logger) {
	resp.Header().Set("Content-Type", "application/x-yaml")
	resp.WriteHeader(this.Status)
	y := yaml.NewEncoder(resp)
	if err := y.Encode(this.Data); err != nil {
		this.logError(resp, req, "Could not render yml response.", err, logProvider)
	}
}

func (this GenericResponse) StreamXmlTo(resp http.ResponseWriter, req *http.Request, logProvider func() log.Logger) {
	resp.Header().Set("Content-Type", "application/xml")
	resp.WriteHeader(this.Status)
	x := xml.NewEncoder(resp)
	x.Indent("", "  ")
	if err := x.Encode(this.Data); err != nil {
		this.logError(resp, req, "Could not render xml response.", err, logProvider)
	}
}

func (this GenericResponse) logError(resp http.ResponseWriter, req *http.Request, message string, err error, logProvider func() log.Logger) {
	if f := this.ErrorHandler; f != nil {
		f(resp, req, message, err, this.Status)
		return
	}
	logProvider().
		WithAll(map[string]any{
			"runtime":       Runtime(),
			"requestId":     this.RequestId,
			"correlationId": this.CorrelationId,
			"remoteIp":      RequestBasedLazyFor(req, RemoteIpOfRequest),
			"host":          RequestBasedLazyFor(req, HostOfRequest),
			"method":        req.Method,
			"requestUri":    RequestBasedLazyFor(req, UriOfRequest),
			"userAgent":     RequestBasedLazyFor(req, UserAgentOfRequestAny),
		}).
		WithError(err).
		Error(message)
}
