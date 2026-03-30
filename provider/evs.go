package provider

import (
	"context"

	dto "github.com/prometheus/client_model/go"

	"github.com/iits-consulting/otc-prometheus-exporter/otcclient"
)

// EVSProvider collects CES metrics for the OTC Elastic Volume Service.
type EVSProvider struct{}

func (p *EVSProvider) Namespace() string { return "SYS.EVS" }

func (p *EVSProvider) Collect(ctx context.Context, client *otcclient.Client) ([]*dto.MetricFamily, error) {
	return CollectCESMetrics(ctx, client, "SYS.EVS")
}

func (p *EVSProvider) Dashboard() DashboardConfig {
	return DashboardConfig{
		Title: "EVS - Elastic Volume Service",
		UID:   "otc-evs",
		Sections: []PanelSection{
			{Title: "Metrics", Panels: []PanelConfig{
				{Metric: "evs_disk_read_bytes_rate", Title: "Read Throughput", Unit: "Bps", Type: TimeSeries},
				{Metric: "evs_disk_write_bytes_rate", Title: "Write Throughput", Unit: "Bps", Type: TimeSeries},
				{Metric: "evs_disk_read_requests_rate", Title: "Read IOPS", Unit: "iops", Type: TimeSeries},
				{Metric: "evs_disk_write_requests_rate", Title: "Write IOPS", Unit: "iops", Type: TimeSeries},
			}},
		},
	}
}
