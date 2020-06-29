package context

type MetricsCollector interface {
	CollectContext(*Context)

	CollectClientStarted() func()
	CollectUpstreamStarted() func()
}
