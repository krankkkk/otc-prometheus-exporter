package otcclient

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptrace"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/iits-consulting/otc-prometheus-exporter/internal"
)

// HTTP trace metrics exposed on the exporter self-monitoring endpoint.
var (
	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "otc_http_request_duration_seconds",
			Help:    "Total duration of outgoing HTTP requests to OTC APIs.",
			Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2, 5, 10},
		},
		[]string{"host", "method", "status"},
	)
	HTTPDNSDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "otc_http_dns_duration_seconds",
			Help:    "Duration of DNS lookups for OTC API calls.",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5},
		},
		[]string{"host"},
	)
	HTTPTLSDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "otc_http_tls_duration_seconds",
			Help:    "Duration of TLS handshakes for OTC API calls.",
			Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"host"},
	)
	HTTPTTFBDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "otc_http_ttfb_duration_seconds",
			Help:    "Time to first byte (server processing time) for OTC API calls.",
			Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2, 5},
		},
		[]string{"host", "method"},
	)
	HTTPConnReused = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "otc_http_connections_reused_total",
			Help: "Number of OTC API calls that reused an existing connection.",
		},
		[]string{"host"},
	)
	HTTPConnNew = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "otc_http_connections_new_total",
			Help: "Number of OTC API calls that opened a new connection.",
		},
		[]string{"host"},
	)
)

// instrumentedTransport wraps an http.RoundTripper with httptrace hooks
// to record per-phase timing of each outgoing HTTP request.
type instrumentedTransport struct {
	base   http.RoundTripper
	logger internal.ILogger
}

// newInstrumentedTransport wraps the given transport with HTTP tracing.
func newInstrumentedTransport(base http.RoundTripper, logger internal.ILogger) http.RoundTripper {
	return &instrumentedTransport{base: base, logger: logger}
}

func (t *instrumentedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	host := req.URL.Host
	method := req.Method
	// Extract a short path for logging (strip project ID).
	path := shortPath(req.URL.Path)

	var (
		dnsStart     time.Time
		tlsStart     time.Time
		connectStart time.Time
		reqStart     = time.Now()
	)

	trace := &httptrace.ClientTrace{
		DNSStart: func(_ httptrace.DNSStartInfo) {
			dnsStart = time.Now()
		},
		DNSDone: func(_ httptrace.DNSDoneInfo) {
			if !dnsStart.IsZero() {
				dur := time.Since(dnsStart)
				HTTPDNSDuration.WithLabelValues(host).Observe(dur.Seconds())
				t.logger.Debug("http trace: dns",
					"host", host, "duration", dur.String())
			}
		},
		ConnectStart: func(_, _ string) {
			connectStart = time.Now()
		},
		ConnectDone: func(_, _ string, err error) {
			if !connectStart.IsZero() && err == nil {
				dur := time.Since(connectStart)
				t.logger.Debug("http trace: tcp connect",
					"host", host, "duration", dur.String())
			}
		},
		TLSHandshakeStart: func() {
			tlsStart = time.Now()
		},
		TLSHandshakeDone: func(_ tls.ConnectionState, _ error) {
			if !tlsStart.IsZero() {
				dur := time.Since(tlsStart)
				HTTPTLSDuration.WithLabelValues(host).Observe(dur.Seconds())
				t.logger.Debug("http trace: tls handshake",
					"host", host, "duration", dur.String())
			}
		},
		GotConn: func(info httptrace.GotConnInfo) {
			if info.Reused {
				HTTPConnReused.WithLabelValues(host).Inc()
				t.logger.Debug("http trace: connection reused",
					"host", host)
			} else {
				HTTPConnNew.WithLabelValues(host).Inc()
				t.logger.Debug("http trace: new connection",
					"host", host)
			}
		},
		GotFirstResponseByte: func() {
			dur := time.Since(reqStart)
			HTTPTTFBDuration.WithLabelValues(host, method).Observe(dur.Seconds())
			t.logger.Debug("http trace: first byte",
				"host", host, "method", method, "path", path,
				"duration", dur.String())
		},
	}

	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	// The gophertelekomcloud SDK sets req.Close=true on every request,
	// which disables HTTP keep-alive and forces a new TCP+TLS connection
	// per request. Override this to enable connection reuse.
	req.Close = false

	resp, err := t.base.RoundTrip(req)

	totalDur := time.Since(reqStart)
	status := "error"
	if resp != nil {
		status = fmt.Sprintf("%d", resp.StatusCode)
	}

	HTTPRequestDuration.WithLabelValues(host, method, status).Observe(totalDur.Seconds())

	t.logger.Debug("http trace: complete",
		"host", host, "method", method, "path", path,
		"status", status, "duration", totalDur.String())

	return resp, err
}

// shortPath strips the /V1.0/{project_id}/ or /v1/{project_id}/ prefix
// from API paths for cleaner logging, e.g. "/alarms", "/batch-query-metric-data".
func shortPath(path string) string {
	parts := strings.SplitN(path, "/", 5) // ["", "V1.0", "{project_id}", "rest..."]
	if len(parts) >= 5 {
		return "/" + parts[4]
	}
	return path
}
