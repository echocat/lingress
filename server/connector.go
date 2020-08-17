package server

import (
	"context"
	"net"
	"net/http"
)

const (
	DefaultConnectorIdHttp  = ConnectorId("http")
	DefaultConnectorIdHttps = ConnectorId("https")

	ConnectorKey = "lingress.server.connector"
)

type Connector interface {
	GetId() ConnectorId
}

func GetConnectorOfContext(ctx context.Context) Connector {
	if v, ok := ctx.Value(ConnectorKey).(Connector); ok {
		return v
	} else {
		return nil
	}
}

func ContextWithConnector(ctx context.Context, connector Connector) context.Context {
	return context.WithValue(ctx, ConnectorKey, connector)
}

type ConnectorHandler interface {
	ServeHTTP(Connector, http.ResponseWriter, *http.Request)
	OnConnState(Connector, net.Conn, http.ConnState)
}
