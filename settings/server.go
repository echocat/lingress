package settings

import (
	"fmt"
	"github.com/echocat/lingress/support"
	"github.com/echocat/lingress/value"
)

func NewServer() (Server, error) {
	http, err := NewServerConnector("http", ":8080")
	if err != nil {
		return Server{}, err
	}
	https, err := NewServerConnector("https", ":8443")
	if err != nil {
		return Server{}, err
	}
	return Server{
		Http:  http,
		Https: https,

		BehindReverseProxy: value.False(),
	}, nil
}

type Server struct {
	Http               ServerConnector `yaml:"http,omitempty" json:"http,omitempty"`
	Https              ServerConnector `yaml:"https,omitempty" json:"https,omitempty"`
	BehindReverseProxy value.Bool      `yaml:"behindReverseProxy,omitempty" json:"behindReverseProxy,omitempty"`
}

func (this *Server) RegisterFlags(fe support.FlagEnabled, appPrefix string) {
	this.Http.RegisterFlags(fe, appPrefix)
	this.Https.RegisterFlags(fe, appPrefix)

	fe.Flag("server.behindReverseProxy", "If true also X-Forwarded headers are evaluated before send to upstream.").
		PlaceHolder(this.BehindReverseProxy.String()).
		Envar(support.FlagEnvName(appPrefix, "SERVER_BEHIND_REVERSE_PROXY")).
		SetValue(&this.BehindReverseProxy)
}

func (this *Server) GetById(id string) (*ServerConnector, error) {
	switch id {
	case "http":
		return &this.Http, nil
	case "https":
		return &this.Https, nil
	default:
		return nil, fmt.Errorf("don't know server kind %q", id)
	}
}
