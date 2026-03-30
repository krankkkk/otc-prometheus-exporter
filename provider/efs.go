package provider

import (
	"context"

	dto "github.com/prometheus/client_model/go"

	"github.com/iits-consulting/otc-prometheus-exporter/otcclient"
)

// EFSProvider collects CES metrics for the OTC Elastic File Service.
type EFSProvider struct{}

func (p *EFSProvider) Namespace() string { return "SYS.EFS" }

func (p *EFSProvider) Collect(ctx context.Context, client *otcclient.Client) ([]*dto.MetricFamily, error) {
	return CollectCESMetrics(ctx, client, "SYS.EFS")
}

func (p *EFSProvider) Dashboard() DashboardConfig {
	return DashboardConfig{
		Title: "EFS - SFS Turbo",
		UID:   "otc-efs",
		Sections: []PanelSection{
			{Title: "Metrics", Panels: []PanelConfig{
				{Metric: "efs_read_bandwidth", Title: "Read Bandwidth", Unit: "Bps", Type: TimeSeries},
				{Metric: "efs_write_bandwidth", Title: "Write Bandwidth", Unit: "Bps", Type: TimeSeries},
			}},
		},
	}
}
