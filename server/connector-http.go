package server

import (
	"context"
	"github.com/echocat/lingress/settings"
	"github.com/echocat/lingress/support"
	"github.com/echocat/slf4g"
	"github.com/echocat/slf4g/level"
	sdk "github.com/echocat/slf4g/sdk/bridge"
	"github.com/pires/go-proxyproto"
	"net"
	"net/http"
	"time"
)

type HttpConnector struct {
	settings       *settings.Settings
	clientSettings *settings.ClientConnector
	serverSettings *settings.ServerConnector

	Id      ConnectorId
	Handler ConnectorHandler
	Logger  log.Logger

	Server       http.Server
	ListenConfig net.ListenConfig
}

func NewHttpConnector(s *settings.Settings, id ConnectorId, logger log.Logger) (*HttpConnector, error) {
	clientSettings, err := s.Client.GetById(string(id))
	if err != nil {
		return nil, err
	}
	serverSettings, err := s.Server.GetById(string(id))
	if err != nil {
		return nil, err
	}
	result := HttpConnector{
		settings:       s,
		clientSettings: clientSettings,
		serverSettings: serverSettings,

		Id:     id,
		Logger: logger,

		Server: http.Server{
			ErrorLog: sdk.NewWrapper(logger, level.Debug),
		},

		ListenConfig: net.ListenConfig{},
	}

	result.Server.Handler = http.HandlerFunc(result.serveHTTP)
	result.Server.ConnState = result.onConnState

	return &result, nil
}

func (this *HttpConnector) Serve(stop support.Channel) error {
	if err := this.serverSettings.ApplyToHttpServer(&this.Server); err != nil {
		return err
	}
	if err := this.clientSettings.ApplyToHttpServer(&this.Server); err != nil {
		return err
	}
	if err := this.clientSettings.ApplyToNetListenConfig(&this.ListenConfig); err != nil {
		return err
	}

	ln, err := (&this.ListenConfig).Listen(context.Background(), "tcp", this.Server.Addr)
	if err != nil {
		return err
	}
	ln = newLimitedListener(this.serverSettings.MaxConnections, this.serverSettings.SoLinger, ln)

	if this.serverSettings.RespectProxyProtocol {
		ln = &proxyproto.Listener{Listener: ln}
	}

	var serve func() error
	if tlsConfig := this.Server.TLSConfig; tlsConfig != nil {
		serve = func() error {
			return this.Server.ServeTLS(ln, "", "")
		}
	} else {
		serve = func() error {
			return this.Server.Serve(ln)
		}
	}

	go func() {
		if err := serve(); err != nil && err != http.ErrServerClosed {
			this.Logger.
				WithError(err).
				With("address", this.Server.Addr).
				Error("server is unable to serve proxy interface")
			stop.Broadcast()
		}
	}()
	this.Logger.
		With("address", this.Server.Addr).
		Info("serve proxy interface")

	return nil
}

func (this *HttpConnector) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	if err := this.Server.Shutdown(ctx); err != nil {
		this.Logger.
			WithError(err).
			Warnf("cannot graceful shutdown %s proxy interface %s", this.Id, this.Server.Addr)
	}
	cancel()
}

func (this *HttpConnector) GetId() ConnectorId {
	return this.Id
}

func (this *HttpConnector) serveHTTP(resp http.ResponseWriter, req *http.Request) {
	if v := this.Handler; v != nil {
		v.ServeHTTP(this, resp, req)
	}
}

func (this *HttpConnector) onConnState(conn net.Conn, state http.ConnState) {
	if v := this.Handler; v != nil {
		v.OnConnState(this, conn, state)
	}
}
