package provider

import (
	"context"

	dto "github.com/prometheus/client_model/go"

	"github.com/iits-consulting/otc-prometheus-exporter/otcclient"
)

// GaussDBV5Provider collects CES metrics for the OTC GaussDB v5 service.
type GaussDBV5Provider struct{}

func (p *GaussDBV5Provider) Namespace() string { return "SYS.GAUSSDBV5" }

func (p *GaussDBV5Provider) Collect(ctx context.Context, client *otcclient.Client) ([]*dto.MetricFamily, error) {
	return CollectCESMetrics(ctx, client, "SYS.GAUSSDBV5")
}

func (p *GaussDBV5Provider) Dashboard() DashboardConfig {
	return DashboardConfig{
		Title: "GaussDB for openGauss",
		UID:   "otc-gaussdbv5",
		Sections: []PanelSection{
			{Title: "Metrics", Panels: []PanelConfig{
				{Metric: "gaussdbv5_cpu_usage", Title: "CPU Usage", Unit: "percent", Type: TimeSeries,
					Thresholds: []Threshold{{Value: 0, Color: "green"}, {Value: 80, Color: "yellow"}, {Value: 90, Color: "red"}}},
				{Metric: "gaussdbv5_mem_usage", Title: "Memory Usage", Unit: "percent", Type: TimeSeries,
					Thresholds: []Threshold{{Value: 0, Color: "green"}, {Value: 80, Color: "yellow"}, {Value: 90, Color: "red"}}},
			}},
		},
	}
}
