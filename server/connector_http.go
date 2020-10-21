package server

import (
	"context"
	"fmt"
	"github.com/echocat/lingress/support"
	"github.com/pires/go-proxyproto"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"strings"
	"time"
)

type HttpConnector struct {
	Id      ConnectorId
	Handler ConnectorHandler

	SoLinger       int16
	MaxConnections uint16

	// See https://www.haproxy.org/download/2.3/doc/proxy-protocol.txt
	RespectProxyProtocol bool

	Server       http.Server
	ListenConfig net.ListenConfig
}

func NewHttpConnector(id ConnectorId) (*HttpConnector, error) {
	result := HttpConnector{
		Id:             id,
		MaxConnections: 512,
		SoLinger:       -1,

		Server: http.Server{
			Addr:              ":8080",
			MaxHeaderBytes:    2 << 20, // 2MB,
			ReadHeaderTimeout: 30 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       5 * time.Minute,
			ErrorLog: support.StdLog(log.Fields{
				"context": "server.http",
			}, log.DebugLevel),
		},

		ListenConfig: net.ListenConfig{
			KeepAlive: 2 * time.Minute,
		},
	}

	result.Server.Handler = http.HandlerFunc(result.serveHTTP)
	result.Server.ConnState = result.onConnState

	return &result, nil
}

func (instance *HttpConnector) Serve(stop support.Channel) error {
	ln, err := (&instance.ListenConfig).Listen(context.Background(), "tcp", instance.Server.Addr)
	if err != nil {
		return err
	}
	ln = newLimitedListener(instance.MaxConnections, instance.SoLinger, ln)

	if instance.RespectProxyProtocol {
		ln = &proxyproto.Listener{Listener: ln}
	}

	var serve func() error
	if tlsConfig := instance.Server.TLSConfig; tlsConfig != nil {
		serve = func() error {
			return instance.Server.ServeTLS(ln, "", "")
		}
	} else {
		serve = func() error {
			return instance.Server.Serve(ln)
		}
	}

	go func() {
		if err := serve(); err != nil && err != http.ErrServerClosed {
			log.WithError(err).
				WithField("address", instance.Server.Addr).
				Error("server is unable to serve proxy interface")
			stop.Broadcast()
		}
	}()
	log.WithField("address", instance.Server.Addr).
		Info("serve proxy interface")

	return nil
}

func (instance *HttpConnector) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	if err := instance.Server.Shutdown(ctx); err != nil {
		log.WithError(err).
			Warnf("cannot graceful shutdown %s proxy interface %s", instance.Id, instance.Server.Addr)
	}
	cancel()
}

func (instance *HttpConnector) flagName(prefix, suffix string) string {
	return fmt.Sprintf("%s.%s.%s", prefix, instance.Id, suffix)
}

func (instance *HttpConnector) serverFlagName(suffix string) string {
	return instance.flagName("server", suffix)
}

func (instance *HttpConnector) clientFlagName(suffix string) string {
	return instance.flagName("client", suffix)
}

func (instance *HttpConnector) flagEnvVar(appPrefix string, prefix, suffix string) string {
	return support.FlagEnvName(appPrefix, fmt.Sprintf("%s_%s_%s", prefix, strings.ToUpper(string(instance.Id)), suffix))
}

func (instance *HttpConnector) serverFlagEnvVar(appPrefix string, suffix string) string {
	return instance.flagEnvVar(appPrefix, "SERVER", suffix)
}

func (instance *HttpConnector) clientFlagEnvVar(appPrefix string, suffix string) string {
	return instance.flagEnvVar(appPrefix, "CLIENT", suffix)
}

func (instance *HttpConnector) RegisterFlag(fe support.FlagEnabled, appPrefix string) error {
	fe.Flag(instance.serverFlagName("address"), "Listen address where the proxy is listening to serve").
		PlaceHolder(instance.Server.Addr).
		Envar(instance.serverFlagEnvVar(appPrefix, "address")).
		StringVar(&instance.Server.Addr)
	fe.Flag(instance.serverFlagName("maxConnections"), "Maximum amount of connections handled by lingress concurrently via HTTP.").
		PlaceHolder(fmt.Sprint(instance.MaxConnections)).
		Envar(instance.serverFlagEnvVar(appPrefix, "MAX_CONNECTIONS")).
		Uint16Var(&instance.MaxConnections)
	fe.Flag(instance.serverFlagName("soLinger"), "Set the behavior of SO_LINGER.").
		PlaceHolder(fmt.Sprint(instance.SoLinger)).
		Envar(instance.serverFlagEnvVar(appPrefix, "SO_LINGER")).
		Int16Var(&instance.SoLinger)
	fe.Flag(instance.serverFlagName("proxyProtocol.respect"), "If set to true the proxy protocol will be respected. See: https://www.haproxy.org/download/2.3/doc/proxy-protocol.txt").
		PlaceHolder(fmt.Sprint(instance.RespectProxyProtocol)).
		Envar(instance.serverFlagEnvVar(appPrefix, "PROXY_PROTOCOL_RESPECT")).
		BoolVar(&instance.RespectProxyProtocol)

	fe.Flag(instance.clientFlagName("maxHeaderBytes"), "Maximum number of bytes the server will read parsing the request header's keys and values, including the request line. It does not limit the size of the request body.").
		PlaceHolder(fmt.Sprint(instance.Server.MaxHeaderBytes)).
		Envar(instance.clientFlagEnvVar(appPrefix, "MAX_HEADER_BYTES")).
		IntVar(&instance.Server.MaxHeaderBytes)
	fe.Flag(instance.clientFlagName("readHeaderTimeout"), "Amount of time allowed to read request headers. The connection's read deadline is reset after reading the headers and the Handler can decide what is considered too slow for the body.").
		PlaceHolder(fmt.Sprint(instance.Server.ReadHeaderTimeout)).
		Envar(instance.clientFlagEnvVar(appPrefix, "READ_HEADER_TIMEOUT")).
		DurationVar(&instance.Server.ReadHeaderTimeout)
	fe.Flag(instance.clientFlagName("writeTimeout"), "Maximum duration before timing out writes of the response. It is reset whenever a new request's header is read.").
		PlaceHolder(fmt.Sprint(instance.Server.WriteTimeout)).
		Envar(instance.clientFlagEnvVar(appPrefix, "WRITE_TIMEOUT")).
		DurationVar(&instance.Server.WriteTimeout)
	fe.Flag(instance.clientFlagName("idleTimeout"), "Maximum amount of time to wait for the next request when keep-alives are enabled.").
		PlaceHolder(fmt.Sprint(instance.Server.IdleTimeout)).
		Envar(instance.clientFlagEnvVar(appPrefix, "IDLE_TIMEOUT")).
		DurationVar(&instance.Server.IdleTimeout)
	fe.Flag(instance.clientFlagName("keepAlive"), "Duration to keep a connection alive (if required); 0 means unlimited.").
		PlaceHolder(fmt.Sprint(instance.ListenConfig.KeepAlive)).
		Envar(instance.clientFlagEnvVar(appPrefix, "KEEP_ALIVE")).
		DurationVar(&instance.ListenConfig.KeepAlive)

	return nil
}

func (instance *HttpConnector) GetId() ConnectorId {
	return instance.Id
}

func (instance *HttpConnector) serveHTTP(resp http.ResponseWriter, req *http.Request) {
	if v := instance.Handler; v != nil {
		v.ServeHTTP(instance, resp, req)
	}
}

func (instance *HttpConnector) onConnState(conn net.Conn, state http.ConnState) {
	if v := instance.Handler; v != nil {
		v.OnConnState(instance, conn, state)
	}
}
