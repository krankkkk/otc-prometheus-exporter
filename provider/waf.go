package provider

import (
	"context"

	dto "github.com/prometheus/client_model/go"

	"github.com/iits-consulting/otc-prometheus-exporter/otcclient"
)

// WAFProvider collects CES metrics for the OTC Web Application Firewall service.
type WAFProvider struct{}

func (p *WAFProvider) Namespace() string { return "SYS.WAF" }

func (p *WAFProvider) Collect(ctx context.Context, client *otcclient.Client) ([]*dto.MetricFamily, error) {
	return CollectCESMetrics(ctx, client, "SYS.WAF")
}

func (p *WAFProvider) Dashboard() DashboardConfig {
	return DashboardConfig{
		Title: "WAF - Web Application Firewall",
		UID:   "otc-waf",
		Sections: []PanelSection{
			{Title: "Metrics", Panels: []PanelConfig{
				{Metric: "waf_requests", Title: "Requests", Unit: "short", Type: TimeSeries},
				{Metric: "waf_waf_http_code_2xx", Title: "2xx Responses", Unit: "short", Type: TimeSeries},
				{Metric: "waf_waf_http_code_4xx", Title: "4xx Responses", Unit: "short", Type: TimeSeries},
				{Metric: "waf_waf_http_code_5xx", Title: "5xx Responses", Unit: "short", Type: TimeSeries,
					Thresholds: []Threshold{{Value: 0, Color: "green"}, {Value: 1, Color: "red"}}},
			}},
		},
	}
}
