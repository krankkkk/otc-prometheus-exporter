package provider

import (
	"context"

	vpcBW "github.com/opentelekomcloud/gophertelekomcloud/openstack/vpc/v1/bandwidths"
	dto "github.com/prometheus/client_model/go"

	"github.com/iits-consulting/otc-prometheus-exporter/otcclient"
)

// VPCProvider collects CES metrics for the OTC Virtual Private Cloud service,
// enriches them with bandwidth names, and reports bandwidth size.
type VPCProvider struct{}

func (p *VPCProvider) Namespace() string { return "SYS.VPC" }

func (p *VPCProvider) Collect(ctx context.Context, client *otcclient.Client) ([]*dto.MetricFamily, error) {
	return CollectWithEnrichment(ctx, client, "SYS.VPC", func(ctx context.Context, client *otcclient.Client) (*EnrichResult, error) {
		vpcClient, err := client.NetworkV1()
		if err != nil {
			return nil, err
		}
		pages, err := vpcBW.List(vpcClient, vpcBW.ListOpts{}).AllPages()
		if err != nil {
			return nil, err
		}
		bandwidths, err := vpcBW.ExtractBandWidths(pages)
		if err != nil {
			return nil, err
		}
		return &EnrichResult{
			NameMap:       buildVPCNameMap(bandwidths),
			ExtraFamilies: convertVPCBandwidthsToMetrics(bandwidths),
		}, nil
	})
}

// buildVPCNameMap creates a mapping from bandwidth ID to bandwidth name.
func buildVPCNameMap(bandwidths []vpcBW.BandWidth) map[string]string {
	m := make(map[string]string, len(bandwidths))
	for _, bw := range bandwidths {
		m[bw.ID] = bw.Name
	}
	return m
}

// convertVPCBandwidthsToMetrics creates a MetricFamily "vpc_bandwidth_size_mbit"
// with a gauge metric per bandwidth reporting its configured size in Mbit/s.
func convertVPCBandwidthsToMetrics(bandwidths []vpcBW.BandWidth) []*dto.MetricFamily {
	metrics := make([]*dto.Metric, 0, len(bandwidths))
	for _, bw := range bandwidths {
		metrics = append(metrics, NewGaugeMetric(float64(bw.Size), map[string]string{
			"resource_id":   bw.ID,
			"resource_name": bw.Name,
		}))
	}
	return []*dto.MetricFamily{NewGaugeMetricFamily("vpc_bandwidth_size_mbit", metrics)}
}

