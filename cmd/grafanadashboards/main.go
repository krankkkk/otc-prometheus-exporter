package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/iits-consulting/otc-prometheus-exporter/grafana"
	"github.com/iits-consulting/otc-prometheus-exporter/provider"
	"github.com/spf13/cobra"
)

func allDashboardProviders() []provider.DashboardProvider {
	providers := []provider.MetricProvider{
		&provider.ECSProvider{},
		&provider.RDSProvider{},
		&provider.ELBProvider{},
		&provider.DMSProvider{},
		&provider.NATProvider{},
		&provider.DCSProvider{},
		&provider.DDSProvider{},
		&provider.CBRProvider{},
		&provider.ASProvider{},
		&provider.VPCProvider{},
		&provider.OBSProvider{},
		&provider.WAFProvider{},
		&provider.BMSProvider{},
		&provider.EVSProvider{},
		&provider.SFSProvider{},
		&provider.EFSProvider{},
		&provider.DWSProvider{},
		&provider.CSSProvider{},
		&provider.GaussDBProvider{},
		&provider.GaussDBV5Provider{},
		&provider.NoSQLProvider{},
		&provider.VPNProvider{},
		&provider.AlarmProvider{},
	}

	var dashProviders []provider.DashboardProvider
	for _, p := range providers {
		if dp, ok := p.(provider.DashboardProvider); ok {
			dashProviders = append(dashProviders, dp)
		}
	}
	return dashProviders
}

// exporterDashboard returns the dashboard config for the exporter's own
// instrumentation metrics (scrape stats, Go runtime, process info).
func exporterDashboard() provider.DashboardConfig {
	return provider.DashboardConfig{
		Title: "OTC Exporter",
		UID:   "otc-exporter",
		Sections: []provider.PanelSection{
			{Title: "Scrape Duration", Panels: []provider.PanelConfig{
				{Metric: "otc_scrape_duration_seconds_bucket", Title: "Scrape Duration p50", Unit: "s", Type: provider.TimeSeries,
					Expr:   `histogram_quantile(0.50, sum by (namespace, le) (rate(otc_scrape_duration_seconds_bucket[5m])))`,
					Legend: "{{namespace}}"},
				{Metric: "otc_scrape_duration_seconds_bucket", Title: "Scrape Duration p95", Unit: "s", Type: provider.TimeSeries,
					Expr:   `histogram_quantile(0.95, sum by (namespace, le) (rate(otc_scrape_duration_seconds_bucket[5m])))`,
					Legend: "{{namespace}}"},
			}},
			{Title: "Scrape Results", Panels: []provider.PanelConfig{
				{Metric: "otc_scrape_families_count", Title: "Families per Namespace", Unit: "short", Type: provider.TimeSeries,
					Legend: "{{namespace}}"},
				{Metric: "otc_scrape_metrics_count", Title: "Metrics per Namespace", Unit: "short", Type: provider.TimeSeries,
					Legend: "{{namespace}}"},
			}},
			{Title: "HTTP Request Duration", Panels: []provider.PanelConfig{
				{Metric: "otc_http_request_duration_seconds_bucket", Title: "Request Duration p50", Unit: "s", Type: provider.TimeSeries,
					Expr:   `histogram_quantile(0.50, sum by (host, le) (rate(otc_http_request_duration_seconds_bucket[5m])))`,
					Legend: "{{host}}"},
				{Metric: "otc_http_request_duration_seconds_bucket", Title: "Request Duration p95", Unit: "s", Type: provider.TimeSeries,
					Expr:   `histogram_quantile(0.95, sum by (host, le) (rate(otc_http_request_duration_seconds_bucket[5m])))`,
					Legend: "{{host}}"},
			}},
			{Title: "HTTP Connection Phases", Panels: []provider.PanelConfig{
				{Metric: "otc_http_dns_duration_seconds_bucket", Title: "DNS Lookup p95", Unit: "s", Type: provider.TimeSeries,
					Expr:   `histogram_quantile(0.95, sum by (host, le) (rate(otc_http_dns_duration_seconds_bucket[5m])))`,
					Legend: "{{host}}"},
				{Metric: "otc_http_tls_duration_seconds_bucket", Title: "TLS Handshake p95", Unit: "s", Type: provider.TimeSeries,
					Expr:   `histogram_quantile(0.95, sum by (host, le) (rate(otc_http_tls_duration_seconds_bucket[5m])))`,
					Legend: "{{host}}"},
				{Metric: "otc_http_ttfb_duration_seconds_bucket", Title: "Time to First Byte p95", Unit: "s", Type: provider.TimeSeries,
					Expr:   `histogram_quantile(0.95, sum by (host, method, le) (rate(otc_http_ttfb_duration_seconds_bucket[5m])))`,
					Legend: "{{host}} {{method}}"},
			}},
			{Title: "HTTP Connections", Panels: []provider.PanelConfig{
				{Metric: "otc_http_connections_reused_total", Title: "Connections Reused", Unit: "ops", Type: provider.TimeSeries,
					Expr:   `rate(otc_http_connections_reused_total[5m])`,
					Legend: "{{host}}"},
				{Metric: "otc_http_connections_new_total", Title: "New Connections", Unit: "ops", Type: provider.TimeSeries,
					Expr:   `rate(otc_http_connections_new_total[5m])`,
					Legend: "{{host}}"},
			}},
			{Title: "Go Runtime", Panels: []provider.PanelConfig{
				{Metric: "go_goroutines", Title: "Goroutines", Unit: "short", Type: provider.TimeSeries},
				{Metric: "go_memstats_alloc_bytes", Title: "Memory Allocated", Unit: "bytes", Type: provider.TimeSeries},
				{Metric: "go_gc_duration_seconds", Title: "GC Duration", Unit: "s", Type: provider.TimeSeries},
			}},
			{Title: "Process", Panels: []provider.PanelConfig{
				{Metric: "process_cpu_seconds_total", Title: "CPU Usage", Unit: "s", Type: provider.TimeSeries},
				{Metric: "process_resident_memory_bytes", Title: "Resident Memory", Unit: "bytes", Type: provider.TimeSeries},
				{Metric: "process_open_fds", Title: "Open File Descriptors", Unit: "short", Type: provider.TimeSeries},
			}},
		},
	}
}

func main() {
	var outputPath string
	var rootCmd = &cobra.Command{
		Use:   "grafanadashboards",
		Short: "Generates Grafana dashboards from provider dashboard metadata.",
		Run: func(cmd *cobra.Command, args []string) {
			if err := os.MkdirAll(outputPath, 0755); err != nil {
				log.Fatalf("Could not create output directory %s: %v\n", outputPath, err)
			}

			// Generate the exporter self-monitoring dashboard.
			allConfigs := []provider.DashboardConfig{exporterDashboard()}
			for _, dp := range allDashboardProviders() {
				allConfigs = append(allConfigs, dp.Dashboard())
			}

			for _, cfg := range allConfigs {
				board := grafana.GenerateDashboard(cfg)

				b, err := json.MarshalIndent(board, "", "  ")
				if err != nil {
					log.Fatalf("Could not marshal dashboard %s: %v\n", cfg.Title, err)
				}

				filename := strings.ToLower(strings.ReplaceAll(cfg.UID, "otc-", "")) + ".json"
				outputFile := path.Join(outputPath, filename)
				if err := os.WriteFile(outputFile, b, 0644); err != nil {
					log.Fatalf("Could not write %s: %v\n", outputFile, err)
				}
				fmt.Printf("Generated %s\n", outputFile)
			}
		},
	}
	rootCmd.Flags().StringVar(&outputPath, "output-path", "", "Directory for generated dashboards.")
	rootCmd.MarkFlagRequired("output-path") //nolint:errcheck

	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
