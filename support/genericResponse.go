package support

import (
	"encoding/json"
	"encoding/xml"
	"gopkg.in/yaml.v2"
	"net/http"
	"time"
)

type GenericResponse struct {
	XMLName   xml.Name    `json:"-" yaml:"-" xml:"response"`
	Timestamp string      `json:"timestamp" yaml:"timestamp" xml:"timestamp"`
	Status    int         `json:"status" yaml:"status" xml:"status"`
	Message   string      `json:"message,omitempty" yaml:"message,omitempty" xml:"message,omitempty"`
	Path      string      `json:"path" yaml:"path" xml:"path"`
	RequestId string      `json:"requestId,omitempty" yaml:"requestId,omitempty" xml:"requestId,omitempty"`
	Data      interface{} `json:"data,omitempty" yaml:"data,omitempty" xml:"data,omitempty"`
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

func (instance GenericResponse) WithUri(uri string) (GenericResponse, error) {
	if p, err := PathOfUri(uri); err != nil {
		return GenericResponse{}, err
	} else {
		return instance.WithPath(p), nil
	}
}

func (instance GenericResponse) WithPath(path string) GenericResponse {
	result := instance
	result.Path = path
	return result
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
	if err := j.Encode(instance); err != nil {
		LogForRequest(req).
			WithError(err).
			WithField("statusCode", instance.Status).
			Error("Could not render json response.")
	}
}

func (instance GenericResponse) StreamYamlTo(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Set("Content-Type", "application/x-yaml")
	resp.WriteHeader(instance.Status)
	y := yaml.NewEncoder(resp)
	if err := y.Encode(instance); err != nil {
		LogForRequest(req).
			WithError(err).
			WithField("statusCode", instance.Status).
			Error("Could not render yml response.")
	}
}

func (instance GenericResponse) StreamXmlTo(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Set("Content-Type", "application/xml")
	resp.WriteHeader(instance.Status)
	x := xml.NewEncoder(resp)
	x.Indent("", "  ")
	if err := x.Encode(instance); err != nil {
		LogForRequest(req).
			WithError(err).
			WithField("statusCode", instance.Status).
			Error("Could not render xml response.")
	}
}
