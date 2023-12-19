package settings

import (
	"fmt"
	"github.com/echocat/lingress/support"
)

func NewClient() (Client, error) {
	http, err := NewClientConnector("http")
	if err != nil {
		return Client{}, err
	}
	https, err := NewClientConnector("https")
	if err != nil {
		return Client{}, err
	}
	return Client{
		Http:  http,
		Https: https,
	}, nil
}

type Client struct {
	Http  ClientConnector `yaml:"http,omitempty" json:"http,omitempty"`
	Https ClientConnector `yaml:"https,omitempty" json:"https,omitempty"`
}

func (this *Client) RegisterFlags(fe support.FlagEnabled, appPrefix string) {
	this.Http.RegisterFlags(fe, appPrefix)
	this.Https.RegisterFlags(fe, appPrefix)
}

func (this *Client) GetById(id string) (*ClientConnector, error) {
	switch id {
	case "http":
		return &this.Http, nil
	case "https":
		return &this.Https, nil
	default:
		return nil, fmt.Errorf("don't know client kind %q", id)
	}
}
