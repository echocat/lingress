package management

import (
	"github.com/echocat/lingress/context"
	"github.com/echocat/lingress/rules"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"strconv"
)

var (
	MetricsLabelNames = []string{
		"client_status",
		"client_status_summary",
		"upstream_status",
		"upstream_status_summary",
		"rule",
	}
)

type Metrics struct {
	Client   *ClientMetrics
	Upstream *UpstreamMetrics
	Rules    *RulesMetrics

	Registry *prometheus.Registry
	Handler  http.Handler
}

type ClientMetrics struct {
	Request *RequestMetrics
}

type UpstreamMetrics struct {
	Request *RequestMetrics
}

type RulesMetrics struct {
	Total   prometheus.GaugeFunc
	Sources prometheus.GaugeFunc

	rules rules.Repository
}

type RequestMetrics struct {
	DurationSeconds *prometheus.HistogramVec
	Total           *prometheus.CounterVec
}

func NewMetrics(rulesRepository rules.Repository) *Metrics {
	registry := prometheus.NewRegistry()

	return &Metrics{
		Client:   NewClientMetrics(registry),
		Upstream: NewUpstreamMetrics(registry),
		Rules:    NewRulesMetrics(registry, rulesRepository),

		Registry: registry,
		Handler:  promhttp.InstrumentMetricHandler(registry, promhttp.HandlerFor(registry, promhttp.HandlerOpts{})),
	}
}

func NewRequestMetrics(registerer prometheus.Registerer, variant string, buckets []float64) *RequestMetrics {
	return &RequestMetrics{
		DurationSeconds: promauto.With(registerer).NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "lingress",
			Subsystem: variant + "_requests",
			Name:      "duration_seconds",
			Help:      "Duration in seconds per request of " + variant + "s.",
			Buckets:   buckets,
		}, MetricsLabelNames),
		Total: promauto.With(registerer).NewCounterVec(prometheus.CounterOpts{
			Namespace: "lingress",
			Subsystem: variant + "_requests",
			Name:      "total",
			Help:      "Amount of requests of " + variant + "s.",
		}, MetricsLabelNames),
	}
}

func NewClientMetrics(registerer prometheus.Registerer) *ClientMetrics {
	return &ClientMetrics{
		Request: NewRequestMetrics(registerer, "client", []float64{
			0.001,
			0.01,
			0.1,
			1,
			10,
		}),
	}
}

func NewUpstreamMetrics(registerer prometheus.Registerer) *UpstreamMetrics {
	return &UpstreamMetrics{
		Request: NewRequestMetrics(registerer, "upstream", []float64{
			0.001,
			0.01,
			0.1,
			1,
			10,
		}),
	}
}

func NewRulesMetrics(registerer prometheus.Registerer, rulesRepository rules.Repository) *RulesMetrics {
	result := &RulesMetrics{
		rules: rulesRepository,
	}

	result.Total = promauto.With(registerer).NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: "lingress",
		Subsystem: "rules",
		Name:      "total",
		Help:      "Total amount of rules handled by lingress.",
	}, result.total)

	result.Sources = promauto.With(registerer).NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: "lingress",
		Subsystem: "rules",
		Name:      "sources",
		Help:      "Total amount of sources (=ingress configurations, ...) of rules handled by lingress.",
	}, result.sources)

	return result
}

func (instance *Metrics) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	instance.Handler.ServeHTTP(resp, req)
}

func (instance *Metrics) Collect(ctx *context.Context) {
	labels := instance.labelsFor(ctx)
	if v := ctx.Client.Duration; v > -1 {
		instance.Client.Request.DurationSeconds.With(labels).Observe(v.Seconds())
		instance.Client.Request.Total.With(labels).Inc()
	}

	if v := ctx.Upstream.Duration; v > -1 {
		instance.Upstream.Request.DurationSeconds.With(labels).Observe(v.Seconds())
		instance.Upstream.Request.Total.With(labels).Inc()
	}
}

func (instance *Metrics) labelsFor(ctx *context.Context) prometheus.Labels {
	result := prometheus.Labels{
		"client_status":         "none",
		"client_status_summary": "none",

		"upstream_status":         "none",
		"upstream_status_summary": "none",

		"rule": "none",
	}

	if v := ctx.Client.Status; v > 0 {
		result["client_status"] = strconv.Itoa(v)
		if v >= 100 && v < 400 {
			result["client_status_summary"] = "ok"
		} else if v < 500 {
			result["client_status_summary"] = "error_client"
		} else if v < 600 {
			result["client_status_summary"] = "error_server"
		}
	}

	if v := ctx.Upstream.Status; v > 0 {
		result["upstream_status"] = strconv.Itoa(v)
		if v >= 100 && v < 400 {
			result["upstream_status_summary"] = "ok"
		} else if v < 500 {
			result["upstream_status_summary"] = "error_client"
		} else if v < 600 {
			result["upstream_status_summary"] = "error_server"
		}
	}

	if v := ctx.Rule; v != nil {
		result["rule"] = v.Source().String()
	}

	return result
}

func (instance *RulesMetrics) total() (result float64) {
	_ = instance.rules.All(func(rules.Rule) error {
		result++
		return nil
	})
	return
}

func (instance *RulesMetrics) sources() float64 {
	result := map[rules.SourceReference]bool{}
	_ = instance.rules.All(func(r rules.Rule) error {
		result[r.Source()] = true
		return nil
	})
	return float64(len(result))
}
