package context

import "github.com/echocat/lingress/server"

type MetricsCollector interface {
	CollectContext(*Context)

	CollectClientStarted(server.ConnectorId) func()
	CollectUpstreamStarted() func()
}
