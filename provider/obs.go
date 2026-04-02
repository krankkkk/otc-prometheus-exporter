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

