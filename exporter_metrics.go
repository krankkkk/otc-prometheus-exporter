package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"

	"github.com/iits-consulting/otc-prometheus-exporter/otcclient"
)

var (
	scrapeDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "otc_scrape_duration_seconds",
			Help:    "Duration of OTC namespace scrape in seconds.",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60},
		},
		[]string{"namespace", "success"},
	)

	scrapeMetrics = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "otc_scrape_metrics_count",
			Help: "Number of metrics returned by the last scrape of a namespace.",
		},
		[]string{"namespace"},
	)

	scrapeFamilies = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "otc_scrape_families_count",
			Help: "Number of metric families returned by the last scrape of a namespace.",
		},
		[]string{"namespace"},
	)
)

// newExporterRegistry creates a Prometheus registry with Go runtime collectors
// and the exporter's own instrumentation metrics.
func newExporterRegistry() *prometheus.Registry {
	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		scrapeDuration,
		scrapeMetrics,
		scrapeFamilies,
		otcclient.HTTPRequestDuration,
		otcclient.HTTPDNSDuration,
		otcclient.HTTPTLSDuration,
		otcclient.HTTPTTFBDuration,
		otcclient.HTTPConnReused,
		otcclient.HTTPConnNew,
	)
	return reg
}
