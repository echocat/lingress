package context

type MetricsCollector interface {
	Collect(*Context)
}
