package provider

import (
	"context"

	dto "github.com/prometheus/client_model/go"

	"github.com/iits-consulting/otc-prometheus-exporter/otcclient"
)

// GaussDBProvider collects CES metrics for the OTC GaussDB service.
type GaussDBProvider struct{}

func (p *GaussDBProvider) Namespace() string { return "SYS.GAUSSDB" }

func (p *GaussDBProvider) Collect(ctx context.Context, client *otcclient.Client) ([]*dto.MetricFamily, error) {
	return CollectCESMetrics(ctx, client, "SYS.GAUSSDB")
}

func (p *GaussDBProvider) Dashboard() DashboardConfig {
	return DashboardConfig{
		Title: "GaussDB for MySQL",
		UID:   "otc-gaussdb",
		Sections: []PanelSection{
			{Title: "Metrics", Panels: []PanelConfig{
				{Metric: "gaussdb_cpu_usage", Title: "CPU Usage", Unit: "percent", Type: TimeSeries,
					Thresholds: []Threshold{{Value: 0, Color: "green"}, {Value: 80, Color: "yellow"}, {Value: 90, Color: "red"}}},
				{Metric: "gaussdb_mem_usage", Title: "Memory Usage", Unit: "percent", Type: TimeSeries,
					Thresholds: []Threshold{{Value: 0, Color: "green"}, {Value: 80, Color: "yellow"}, {Value: 90, Color: "red"}}},
			}},
		},
	}
}
