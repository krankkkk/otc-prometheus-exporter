package provider

import (
	"context"

	dto "github.com/prometheus/client_model/go"

	"github.com/iits-consulting/otc-prometheus-exporter/otcclient"
)

// SFSProvider collects CES metrics for the OTC Scalable File Service.
type SFSProvider struct{}

func (p *SFSProvider) Namespace() string { return "SYS.SFS" }

func (p *SFSProvider) Collect(ctx context.Context, client *otcclient.Client) ([]*dto.MetricFamily, error) {
	return CollectCESMetrics(ctx, client, "SYS.SFS")
}

func (p *SFSProvider) Dashboard() DashboardConfig {
	return DashboardConfig{
		Title: "SFS - Scalable File Service",
		UID:   "otc-sfs",
		Sections: []PanelSection{
			{Title: "Metrics", Panels: []PanelConfig{
				{Metric: "sfs_read_bandwidth", Title: "Read Bandwidth", Unit: "Bps", Type: TimeSeries},
				{Metric: "sfs_write_bandwidth", Title: "Write Bandwidth", Unit: "Bps", Type: TimeSeries},
			}},
		},
	}
}
