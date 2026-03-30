package provider

import (
	"context"

	dto "github.com/prometheus/client_model/go"

	"github.com/iits-consulting/otc-prometheus-exporter/otcclient"
)

// BMSProvider collects CES metrics for the OTC Bare Metal Server service.
type BMSProvider struct{}

func (p *BMSProvider) Namespace() string { return "SERVICE.BMS" }

func (p *BMSProvider) Collect(ctx context.Context, client *otcclient.Client) ([]*dto.MetricFamily, error) {
	return CollectCESMetrics(ctx, client, "SERVICE.BMS")
}

func (p *BMSProvider) Dashboard() DashboardConfig {
	return DashboardConfig{
		Title: "BMS - Bare Metal Server",
		UID:   "otc-bms",
		Sections: []PanelSection{
			{Title: "Metrics", Panels: []PanelConfig{
				{Metric: "bms_cpuusage", Title: "CPU Usage", Unit: "percent", Type: TimeSeries,
					Thresholds: []Threshold{{Value: 0, Color: "green"}, {Value: 80, Color: "yellow"}, {Value: 90, Color: "red"}}},
				{Metric: "bms_memusage", Title: "Memory Usage", Unit: "percent", Type: TimeSeries,
					Thresholds: []Threshold{{Value: 0, Color: "green"}, {Value: 80, Color: "yellow"}, {Value: 90, Color: "red"}}},
				{Metric: "bms_diskreadrate", Title: "Disk Read Rate", Unit: "Bps", Type: TimeSeries},
				{Metric: "bms_diskwriterate", Title: "Disk Write Rate", Unit: "Bps", Type: TimeSeries},
			}},
		},
	}
}
