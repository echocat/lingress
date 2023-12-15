package settings

import (
	"fmt"
	"github.com/echocat/lingress/support"
	"net/http"
	"time"
)

func NewManagement() (Management, error) {
	return Management{
		ListenAddress:         ":8090",
		MaxRequestHeaderBytes: 2 << 20, // 2MB
		ReadHeaderTimeout:     30 * time.Second,
		WriteTimeout:          1 * time.Minute,
		IdleTimeout:           5 * time.Minute,

		Pprof: false,
	}, nil
}

type Management struct {
	ListenAddress         string        `json:"listenAddress,omitempty" yaml:"listenAddress,omitempty"`
	MaxRequestHeaderBytes uint32        `json:"maxRequestHeaderBytes,omitempty" yaml:"maxRequestHeaderBytes,omitempty"`
	ReadHeaderTimeout     time.Duration `json:"readHeaderTimeout,omitempty" yaml:"readHeaderTimeout,omitempty"`
	WriteTimeout          time.Duration `json:"writeTimeout,omitempty" yaml:"writeTimeout,omitempty"`
	IdleTimeout           time.Duration `json:"idleTimeout,omitempty" yaml:"idleTimeout,omitempty"`
	Pprof                 bool          `json:"pprof,omitempty" yaml:"pprof,omitempty"`
}

func (this *Management) RegisterFlags(fe support.FlagEnabled, appPrefix string) {
	fe.Flag("management.listenAddress", "Listen address where the management interface is listening to serve").
		PlaceHolder(this.ListenAddress).
		Envar(support.FlagEnvName(appPrefix, "MANAGEMENT_LISTEN_ADDRESS")).
		StringVar(&this.ListenAddress)
	fe.Flag("management.maxRequestHeaderBytes", "Maximum number of bytes the server will read parsing the request header's keys and values, including the request line. It does not limit the size of the request body.").
		PlaceHolder(fmt.Sprint(this.MaxRequestHeaderBytes)).
		Envar(support.FlagEnvName(appPrefix, "MANAGEMENT_MAX_REQUEST_HEADER_BYTES")).
		Uint32Var(&this.MaxRequestHeaderBytes)
	fe.Flag("management.readHeaderTimeout", "Amount of time allowed to read request headers. The connection's read deadline is reset after reading the headers and the Handler can decide what is considered too slow for the body.").
		PlaceHolder(fmt.Sprint(this.ReadHeaderTimeout)).
		Envar(support.FlagEnvName(appPrefix, "MANAGEMENT_READ_HEADER_TIMEOUT")).
		DurationVar(&this.ReadHeaderTimeout)
	fe.Flag("management.writeTimeout", "Maximum duration before timing out writes of the response. It is reset whenever a new request's header is read.").
		PlaceHolder(fmt.Sprint(this.WriteTimeout)).
		Envar(support.FlagEnvName(appPrefix, "MANAGEMENT_WRITE_TIMEOUT")).
		DurationVar(&this.WriteTimeout)
	fe.Flag("management.idleTimeout", "Maximum amount of time to wait for the next request when keep-alives are enabled.").
		PlaceHolder(fmt.Sprint(this.IdleTimeout)).
		Envar(support.FlagEnvName(appPrefix, "MANAGEMENT_IDLE_TIMEOUT")).
		DurationVar(&this.IdleTimeout)
	fe.Flag("management.pprof", "Will serve at the management endpoint pprof profiling, too.").
		Envar(support.FlagEnvName(appPrefix, "MANAGEMENT_PPROF")).
		BoolVar(&this.Pprof)
}

func (this *Management) ApplyToHttpServer(target *http.Server) error {
	target.Addr = this.ListenAddress
	target.MaxHeaderBytes = int(this.MaxRequestHeaderBytes)
	target.ReadHeaderTimeout = this.ReadHeaderTimeout
	target.WriteTimeout = this.WriteTimeout
	target.IdleTimeout = this.IdleTimeout
	return nil
}
