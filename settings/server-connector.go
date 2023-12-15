package settings

import (
	"fmt"
	"github.com/echocat/lingress/support"
	"net/http"
	"strings"
)

func NewServerConnector(id string, listenAddress string) (ServerConnector, error) {
	maxConnections := uint16(256)
	if id == "https" {
		maxConnections = 512
	}
	return ServerConnector{
		id: id,

		ListenAddress:        listenAddress,
		MaxConnections:       maxConnections,
		SoLinger:             -1,
		RespectProxyProtocol: false,
	}, nil
}

type ServerConnector struct {
	id string

	ListenAddress  string `json:"listenAddress,omitempty" yaml:"listenAddress,omitempty"`
	MaxConnections uint16 `json:"maxConnections,omitempty" yaml:"maxConnections,omitempty"`
	SoLinger       int16  `json:"soLinger,omitempty" yaml:"soLinger,omitempty"`

	// See https://www.haproxy.org/download/2.3/doc/proxy-protocol.txt
	RespectProxyProtocol bool `json:"respectProxyProtocol,omitempty" yaml:"respectProxyProtocol,omitempty"`
}

func (this *ServerConnector) RegisterFlags(fe support.FlagEnabled, appPrefix string) {
	fe.Flag(this.flagName("listenAddress"), "Listen address where the proxy is listening to serve").
		PlaceHolder(this.ListenAddress).
		Envar(this.flagEnvVar(appPrefix, "LISTEN_ADDRESS")).
		StringVar(&this.ListenAddress)
	fe.Flag(this.flagName("maxConnections"), "Maximum amount of connections handled by lingress concurrently via HTTP.").
		PlaceHolder(fmt.Sprint(this.MaxConnections)).
		Envar(this.flagEnvVar(appPrefix, "MAX_CONNECTIONS")).
		Uint16Var(&this.MaxConnections)
	fe.Flag(this.flagName("soLinger"), "Set the behavior of SO_LINGER.").
		PlaceHolder(fmt.Sprint(this.SoLinger)).
		Envar(this.flagEnvVar(appPrefix, "SO_LINGER")).
		Int16Var(&this.SoLinger)
	fe.Flag(this.flagName("proxyProtocol.respect"), "If set to true the proxy protocol will be respected. See: https://www.haproxy.org/download/2.3/doc/proxy-protocol.txt").
		PlaceHolder(fmt.Sprint(this.RespectProxyProtocol)).
		Envar(this.flagEnvVar(appPrefix, "PROXY_PROTOCOL_RESPECT")).
		BoolVar(&this.RespectProxyProtocol)
}

func (this *ServerConnector) flagEnvVar(appPrefix string, suffix string) string {
	return support.FlagEnvName(appPrefix, fmt.Sprintf("SERVER_%s_%s", strings.ToUpper(this.id), suffix))
}

func (this *ServerConnector) flagName(suffix string) string {
	return fmt.Sprintf("server.%s.%s", this.id, suffix)
}

func (this *ServerConnector) ApplyToHttpServer(target *http.Server) error {
	target.Addr = this.ListenAddress
	return nil
}
