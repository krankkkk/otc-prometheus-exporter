package provider

import (
	"context"

	dto "github.com/prometheus/client_model/go"

	"github.com/iits-consulting/otc-prometheus-exporter/otcclient"
)

// CSSProvider collects CES metrics for the OTC Cloud Search Service (Elasticsearch).
type CSSProvider struct{}

func (p *CSSProvider) Namespace() string { return "SYS.ES" }

func (p *CSSProvider) Collect(ctx context.Context, client *otcclient.Client) ([]*dto.MetricFamily, error) {
	return CollectCESMetrics(ctx, client, "SYS.ES")
}

func (p *CSSProvider) Dashboard() DashboardConfig {
	return DashboardConfig{
		Title: "CSS - Cloud Search Service",
		UID:   "otc-css",
		Sections: []PanelSection{
			{Title: "Metrics", Panels: []PanelConfig{
				{Metric: "es_status", Title: "Cluster Status", Unit: "short", Type: Stat,
					Thresholds: []Threshold{{Value: 0, Color: "green"}, {Value: 1, Color: "red"}}},
			}},
		},
	}
}
