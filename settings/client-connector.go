package settings

import (
	"fmt"
	"github.com/echocat/lingress/support"
	"net"
	"net/http"
	"strings"
	"time"
)

func NewClientConnector(id string) (ClientConnector, error) {
	return ClientConnector{
		id: id,

		MaxRequestHeaderBytes: 2 << 20, // 2MB,
		ReadHeaderTimeout:     30 * time.Second,
		WriteTimeout:          30 * time.Second,
		IdleTimeout:           5 * time.Minute,
		KeepAlive:             2 * time.Minute,
	}, nil
}

type ClientConnector struct {
	id string

	MaxRequestHeaderBytes uint32        `json:"maxRequestHeaderBytes,omitempty" yaml:"maxRequestHeaderBytes,omitempty"`
	ReadHeaderTimeout     time.Duration `json:"readHeaderTimeout,omitempty" yaml:"readHeaderTimeout,omitempty"`
	WriteTimeout          time.Duration `json:"writeTimeout,omitempty" yaml:"writeTimeout,omitempty"`
	IdleTimeout           time.Duration `json:"idleTimeout,omitempty" yaml:"idleTimeout,omitempty"`
	KeepAlive             time.Duration `json:"keepAlive,omitempty" yaml:"keepAlive,omitempty"`
}

func (this *ClientConnector) RegisterFlags(fe support.FlagEnabled, appPrefix string) {
	fe.Flag(this.flagName("maxRequestHeaderBytes"), "Maximum number of bytes the client will read parsing the request header's keys and values, including the request line. It does not limit the size of the request body.").
		PlaceHolder(fmt.Sprint(this.MaxRequestHeaderBytes)).
		Envar(this.flagEnvVar(appPrefix, "MAX_REQUEST_HEADER_BYTES")).
		Uint32Var(&this.MaxRequestHeaderBytes)
	fe.Flag(this.flagName("readHeaderTimeout"), "Amount of time allowed to read request headers. The connection's read deadline is reset after reading the headers and the Handler can decide what is considered too slow for the body.").
		PlaceHolder(fmt.Sprint(this.ReadHeaderTimeout)).
		Envar(this.flagEnvVar(appPrefix, "READ_HEADER_TIMEOUT")).
		DurationVar(&this.ReadHeaderTimeout)
	fe.Flag(this.flagName("writeTimeout"), "Maximum duration before timing out writes of the response. It is reset whenever a new request's header is read.").
		PlaceHolder(fmt.Sprint(this.WriteTimeout)).
		Envar(this.flagEnvVar(appPrefix, "WRITE_TIMEOUT")).
		DurationVar(&this.WriteTimeout)
	fe.Flag(this.flagName("idleTimeout"), "Maximum amount of time to wait for the next request when keep-alives are enabled.").
		PlaceHolder(fmt.Sprint(this.IdleTimeout)).
		Envar(this.flagEnvVar(appPrefix, "IDLE_TIMEOUT")).
		DurationVar(&this.IdleTimeout)
	fe.Flag(this.flagName("keepAlive"), "Duration to keep a connection alive (if required); 0 means unlimited.").
		PlaceHolder(fmt.Sprint(this.KeepAlive)).
		Envar(this.flagEnvVar(appPrefix, "KEEP_ALIVE")).
		DurationVar(&this.KeepAlive)
}

func (this *ClientConnector) flagEnvVar(appPrefix string, suffix string) string {
	return support.FlagEnvName(appPrefix, fmt.Sprintf("CLIENT_%s_%s", strings.ToUpper(this.id), suffix))
}

func (this *ClientConnector) flagName(suffix string) string {
	return fmt.Sprintf("client.%s.%s", this.id, suffix)
}

func (this *ClientConnector) ApplyToHttpServer(target *http.Server) error {
	target.MaxHeaderBytes = int(this.MaxRequestHeaderBytes)
	target.ReadHeaderTimeout = this.ReadHeaderTimeout
	target.WriteTimeout = this.WriteTimeout
	target.IdleTimeout = this.IdleTimeout
	return nil
}

func (this *ClientConnector) ApplyToNetListenConfig(target *net.ListenConfig) error {
	target.KeepAlive = this.KeepAlive
	return nil
}
