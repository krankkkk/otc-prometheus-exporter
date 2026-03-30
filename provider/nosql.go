package provider

import (
	"context"

	dto "github.com/prometheus/client_model/go"

	"github.com/iits-consulting/otc-prometheus-exporter/otcclient"
)

// NoSQLProvider collects CES metrics for the OTC NoSQL service.
type NoSQLProvider struct{}

func (p *NoSQLProvider) Namespace() string { return "SYS.NoSQL" }

func (p *NoSQLProvider) Collect(ctx context.Context, client *otcclient.Client) ([]*dto.MetricFamily, error) {
	return CollectCESMetrics(ctx, client, "SYS.NoSQL")
}

func (p *NoSQLProvider) Dashboard() DashboardConfig {
	return DashboardConfig{
		Title: "GaussDB NoSQL",
		UID:   "otc-nosql",
		Sections: []PanelSection{
			{Title: "Metrics", Panels: []PanelConfig{
				{Metric: "nosql_cpu_usage", Title: "CPU Usage", Unit: "percent", Type: TimeSeries,
					Thresholds: []Threshold{{Value: 0, Color: "green"}, {Value: 80, Color: "yellow"}, {Value: 90, Color: "red"}}},
				{Metric: "nosql_mem_usage", Title: "Memory Usage", Unit: "percent", Type: TimeSeries,
					Thresholds: []Threshold{{Value: 0, Color: "green"}, {Value: 80, Color: "yellow"}, {Value: 90, Color: "red"}}},
			}},
		},
	}
}
