package provider

import (
	"context"

	dto "github.com/prometheus/client_model/go"

	"github.com/iits-consulting/otc-prometheus-exporter/otcclient"
)

// DWSProvider collects CES metrics for the OTC Data Warehouse Service.
type DWSProvider struct{}

func (p *DWSProvider) Namespace() string { return "SYS.DWS" }

func (p *DWSProvider) Collect(ctx context.Context, client *otcclient.Client) ([]*dto.MetricFamily, error) {
	return CollectCESMetrics(ctx, client, "SYS.DWS")
}

func (p *DWSProvider) Dashboard() DashboardConfig {
	return DashboardConfig{
		Title: "DWS - Data Warehouse Service",
		UID:   "otc-dws",
		Sections: []PanelSection{
			{Title: "Metrics", Panels: []PanelConfig{
				{Metric: "dws_cpu_usage", Title: "CPU Usage", Unit: "percent", Type: TimeSeries,
					Thresholds: []Threshold{{Value: 0, Color: "green"}, {Value: 80, Color: "yellow"}, {Value: 90, Color: "red"}}},
				{Metric: "dws_mem_usage", Title: "Memory Usage", Unit: "percent", Type: TimeSeries,
					Thresholds: []Threshold{{Value: 0, Color: "green"}, {Value: 80, Color: "yellow"}, {Value: 90, Color: "red"}}},
				{Metric: "dws_disk_usage", Title: "Disk Usage", Unit: "percent", Type: TimeSeries,
					Thresholds: []Threshold{{Value: 0, Color: "green"}, {Value: 80, Color: "yellow"}, {Value: 90, Color: "red"}}},
			}},
		},
	}
}
