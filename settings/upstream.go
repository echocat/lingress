package settings

import (
	"fmt"
	"github.com/echocat/lingress/support"
	"net"
	"net/http"
	"time"
)

func NewUpstream() (Upstream, error) {
	return Upstream{
		MaxIdleConnectionsPerHost: 20,
		MaxConnectionsPerHost:     250,
		IdleConnectionTimeout:     1 * time.Minute,
		MaxResponseHeaderBytes:    10 << 20,
		DialTimeout:               10 * time.Second,
		KeepAlive:                 30 * time.Second,

		OverrideHost:   "",
		OverrideScheme: "",
	}, nil
}

type Upstream struct {
	MaxIdleConnectionsPerHost uint32        `json:"maxIdleConnectionsPerHost,omitempty" yaml:"maxIdleConnectionsPerHost,omitempty"`
	MaxConnectionsPerHost     uint32        `json:"maxConnectionsPerHost,omitempty" yaml:"maxConnectionsPerHost,omitempty"`
	IdleConnectionTimeout     time.Duration `json:"idleConnectionTimeout,omitempty" yaml:"idleConnectionTimeout,omitempty"`
	MaxResponseHeaderBytes    uint32        `json:"maxResponseHeaderBytes,omitempty" yaml:"maxResponseHeaderBytes,omitempty"`
	DialTimeout               time.Duration `json:"dialTimeout,omitempty" yaml:"dialTimeout,omitempty"`
	KeepAlive                 time.Duration `json:"keepAlive,omitempty" yaml:"keepAlive,omitempty"`

	OverrideHost   string `json:"overrideHost,omitempty" yaml:"overrideHost,omitempty"`
	OverrideScheme string `json:"overrideScheme,omitempty" yaml:"overrideScheme,omitempty"`
}

func (this *Upstream) RegisterFlags(fe support.FlagEnabled, appPrefix string) {
	fe.Flag("upstream.maxIdleConnectionsPerHost", "Controls the maximum idle (keep-alive) connections to keep per-host.").
		PlaceHolder(fmt.Sprint(this.MaxIdleConnectionsPerHost)).
		Envar(support.FlagEnvName(appPrefix, "UPSTREAM_MAX_IDLE_CONNECTIONS_PER_HOST")).
		Uint32Var(&this.MaxIdleConnectionsPerHost)
	fe.Flag("upstream.maxConnectionsPerHost", "Limits the total number of connections per host, including connections in the dialing, active, and idle states. On limit violation, dials will block.").
		PlaceHolder(fmt.Sprint(this.MaxConnectionsPerHost)).
		Envar(support.FlagEnvName(appPrefix, "UPSTREAM_MAX_CONNECTIONS_PER_HOST")).
		Uint32Var(&this.MaxConnectionsPerHost)
	fe.Flag("upstream.idleConnectionTimeout", "Maximum amount of time an idle (keep-alive) connection will remain idle before closing itself. Zero means no limit.").
		PlaceHolder(fmt.Sprint(this.IdleConnectionTimeout)).
		Envar(support.FlagEnvName(appPrefix, "UPSTREAM_IDLE_CONNECTION_TIMEOUT")).
		DurationVar(&this.IdleConnectionTimeout)
	fe.Flag("upstream.maxResponseHeaderSize", "Limit on how many response bytes are allowed in the server's response header.").
		PlaceHolder(fmt.Sprint(this.MaxResponseHeaderBytes)).
		Envar(support.FlagEnvName(appPrefix, "UPSTREAM_MAX_RESPONSE_HEADER_SIZE")).
		Uint32Var(&this.MaxResponseHeaderBytes)
	fe.Flag("upstream.dialTimeout", "Maximum amount of time a dial will wait for a connect to complete. If Deadline is also set, it may fail earlier.").
		PlaceHolder(fmt.Sprint(this.DialTimeout)).
		Envar(support.FlagEnvName(appPrefix, "UPSTREAM_DIAL_TIMEOUT")).
		DurationVar(&this.DialTimeout)
	fe.Flag("upstream.keepAlive", "Keep-alive period for an active network connection. If zero, keep-alives are enabled if supported by the protocol and operating system. Network protocols or operating systems that do not support keep-alives ignore this field. If negative, keep-alives are disabled.").
		PlaceHolder(fmt.Sprint(this.KeepAlive)).
		Envar(support.FlagEnvName(appPrefix, "UPSTREAM_KEEP_ALIVE")).
		DurationVar(&this.KeepAlive)
	fe.Flag("upstream.override.host", "Overrides the target host always with this value.").
		PlaceHolder(this.OverrideHost).
		Envar(support.FlagEnvName(appPrefix, "UPSTREAM_OVERRIDE_HOST")).
		StringVar(&this.OverrideHost)
	fe.Flag("upstream.override.scheme", "Overrides the target scheme always with this value.").
		PlaceHolder(this.OverrideScheme).
		Envar(support.FlagEnvName(appPrefix, "UPSTREAM_OVERRIDE_SCHEME")).
		StringVar(&this.OverrideScheme)
}

func (this *Upstream) ApplyToHttpTransport(target *http.Transport) error {
	target.MaxIdleConnsPerHost = int(this.MaxIdleConnectionsPerHost)
	target.MaxConnsPerHost = int(this.MaxConnectionsPerHost)
	target.IdleConnTimeout = this.IdleConnectionTimeout
	target.MaxResponseHeaderBytes = int64(this.MaxResponseHeaderBytes)
	return nil
}

func (this *Upstream) ApplyToNetDialer(target *net.Dialer) error {
	target.Timeout = this.DialTimeout
	target.KeepAlive = this.KeepAlive
	return nil
}
