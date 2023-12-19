package settings

import (
	"github.com/echocat/lingress/support"
)

func New() (Settings, error) {
	accessLog, err := NewAccessLog()
	if err != nil {
		return Settings{}, err
	}
	client, err := NewClient()
	if err != nil {
		return Settings{}, err
	}
	cors, err := NewCors()
	if err != nil {
		return Settings{}, err
	}
	discovery, err := NewDiscovery()
	if err != nil {
		return Settings{}, err
	}
	fallback, err := NewFallback()
	if err != nil {
		return Settings{}, err
	}
	ingress, err := NewIngress()
	if err != nil {
		return Settings{}, err
	}
	kubernetes, err := NewKubernetes()
	if err != nil {
		return Settings{}, err
	}
	management, err := NewManagement()
	if err != nil {
		return Settings{}, err
	}
	request, err := NewRequest()
	if err != nil {
		return Settings{}, err
	}
	response, err := NewResponse()
	if err != nil {
		return Settings{}, err
	}
	server, err := NewServer()
	if err != nil {
		return Settings{}, err
	}
	tls, err := NewTls()
	if err != nil {
		return Settings{}, err
	}
	upstream, err := NewUpstream()
	if err != nil {
		return Settings{}, err
	}
	return Settings{
		AccessLog:  accessLog,
		Client:     client,
		Cors:       cors,
		Discovery:  discovery,
		Fallback:   fallback,
		Ingress:    ingress,
		Kubernetes: kubernetes,
		Management: management,
		Request:    request,
		Response:   response,
		Server:     server,
		Tls:        tls,
		Upstream:   upstream,
	}, nil
}

func MustNew() Settings {
	result, err := New()
	if err != nil {
		panic(err)
	}
	return result
}

type Settings struct {
	AccessLog  AccessLog  `json:"accessLog,omitempty" yaml:"accessLog,omitempty"`
	Client     Client     `json:"client,omitempty" yaml:"client,omitempty"`
	Cors       Cors       `json:"cors,omitempty" yaml:"cors,omitempty"`
	Discovery  Discovery  `json:"discovery,omitempty" yaml:"discovery,omitempty"`
	Fallback   Fallback   `json:"fallback,omitempty" yaml:"fallback,omitempty"`
	Request    Request    `json:"request,omitempty" yaml:"request,omitempty"`
	Response   Response   `json:"response,omitempty" yaml:"response,omitempty"`
	Ingress    Ingress    `json:"ingress,omitempty" yaml:"ingress,omitempty"`
	Kubernetes Kubernetes `json:"kubernetes,omitempty" yaml:"kubernetes,omitempty"`
	Management Management `json:"management,omitempty" yaml:"management,omitempty"`
	Server     Server     `json:"server,omitempty" yaml:"server,omitempty"`
	Tls        Tls        `json:"tls,omitempty" yaml:"tls,omitempty"`
	Upstream   Upstream   `json:"upstream,omitempty" yaml:"upstream,omitempty"`
}

func (this *Settings) RegisterFlags(fe support.FlagEnabled, appPrefix string) {
	this.AccessLog.RegisterFlags(fe, appPrefix)
	this.Client.RegisterFlags(fe, appPrefix)
	this.Cors.RegisterFlags(fe, appPrefix)
	this.Discovery.RegisterFlags(fe, appPrefix)
	this.Fallback.RegisterFlags(fe, appPrefix)
	this.Ingress.RegisterFlags(fe, appPrefix)
	this.Kubernetes.RegisterFlags(fe, appPrefix)
	this.Management.RegisterFlags(fe, appPrefix)
	this.Request.RegisterFlags(fe, appPrefix)
	this.Response.RegisterFlags(fe, appPrefix)
	this.Server.RegisterFlags(fe, appPrefix)
	this.Tls.RegisterFlags(fe, appPrefix)
	this.Upstream.RegisterFlags(fe, appPrefix)
}
