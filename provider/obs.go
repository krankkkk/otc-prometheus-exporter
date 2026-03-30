package provider

import (
	"context"

	dto "github.com/prometheus/client_model/go"

	"github.com/iits-consulting/otc-prometheus-exporter/otcclient"
)

// OBSProvider collects CES metrics for the OTC Object Storage Service.
// OBS is a global service whose CES metrics are only visible under the
// region-level project (e.g. eu-de), not under a specific project.
type OBSProvider struct{}

func (p *OBSProvider) Namespace() string { return "SYS.OBS" }

func (p *OBSProvider) Collect(ctx context.Context, client *otcclient.Client) ([]*dto.MetricFamily, error) {
	families, err := CollectCESMetricsRegionScoped(ctx, client, "SYS.OBS")
	if err != nil {
		return nil, err
	}

	// OBS CES metrics include a bucket_name dimension -- use it as resource_name.
	FillResourceNameFromLabel(families, "bucket_name")

	return families, nil
}

func (p *OBSProvider) Dashboard() DashboardConfig {
	return DashboardConfig{
		Title: "OBS - Object Storage Service",
		UID:   "otc-obs",
		Sections: []PanelSection{
			{Title: "Requests", Panels: []PanelConfig{
				{Metric: "obs_request_count_get_per_second", Title: "GET Requests/s", Unit: "reqps", Type: TimeSeries},
				{Metric: "obs_request_count_put_per_second", Title: "PUT Requests/s", Unit: "reqps", Type: TimeSeries},
				{Metric: "obs_request_count_monitor_2xx", Title: "2xx Responses", Unit: "short", Type: TimeSeries},
				{Metric: "obs_request_count_monitor_4xx", Title: "4xx Responses", Unit: "short", Type: TimeSeries,
					Thresholds: []Threshold{{Value: 0, Color: "green"}, {Value: 1, Color: "yellow"}}},
				{Metric: "obs_request_count_monitor_5xx", Title: "5xx Responses", Unit: "short", Type: TimeSeries,
					Thresholds: []Threshold{{Value: 0, Color: "green"}, {Value: 1, Color: "red"}}},
			}},
			{Title: "Latency", Panels: []PanelConfig{
				{Metric: "obs_request_size_le_1mb_latency_p95", Title: "Latency p95 (<1MB)", Unit: "ms", Type: TimeSeries,
					Thresholds: []Threshold{{Value: 0, Color: "green"}, {Value: 100, Color: "yellow"}, {Value: 500, Color: "red"}}},
			}},
		},
	}
}
