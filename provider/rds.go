package provider

import (
	"context"

	rdsInstances "github.com/opentelekomcloud/gophertelekomcloud/openstack/rds/v3/instances"
	dto "github.com/prometheus/client_model/go"

	"github.com/iits-consulting/otc-prometheus-exporter/otcclient"
)

// RDSProvider collects CES metrics for the OTC Relational Database Service,
// enriches them with instance names, and reports instance/node status and volume size.
type RDSProvider struct{}

func (p *RDSProvider) Namespace() string { return "SYS.RDS" }

func (p *RDSProvider) Collect(ctx context.Context, client *otcclient.Client) ([]*dto.MetricFamily, error) {
	return CollectWithEnrichment(ctx, client, "SYS.RDS", func(ctx context.Context, client *otcclient.Client) (*EnrichResult, error) {
		rdsClient, err := client.RDSV3()
		if err != nil {
			return nil, err
		}
		resp, err := rdsInstances.List(rdsClient, rdsInstances.ListOpts{})
		if err != nil {
			return nil, err
		}
		return &EnrichResult{
			NameMap:       buildRDSNameMap(resp.Instances),
			ExtraFamilies: convertRDSInstancesToMetrics(resp.Instances),
		}, nil
	})
}

func buildRDSNameMap(instances []rdsInstances.InstanceResponse) map[string]string {
	m := make(map[string]string)
	for _, inst := range instances {
		m[inst.Id] = inst.Name
		// Map node IDs too -- CES reports metrics at node level
		// with IDs like "...no03" that differ from the instance ID "...in03".
		for _, node := range inst.Nodes {
			m[node.Id] = inst.Name + "/" + node.Name
		}
	}
	return m
}

func convertRDSInstancesToMetrics(instances []rdsInstances.InstanceResponse) []*dto.MetricFamily {
	statusMetrics := make([]*dto.Metric, 0, len(instances))
	nodeMetrics := make([]*dto.Metric, 0)
	volumeMetrics := make([]*dto.Metric, 0, len(instances))

	for _, inst := range instances {
		statusValue := 1.0
		if inst.Status == "ACTIVE" {
			statusValue = 0.0
		}
		statusMetrics = append(statusMetrics, NewGaugeMetric(statusValue, map[string]string{
			"resource_id":   inst.Id,
			"resource_name": inst.Name,
			"status":        inst.Status,
		}))

		for _, node := range inst.Nodes {
			nodeValue := 1.0
			if node.Status == "ACTIVE" {
				nodeValue = 0.0
			}
			nodeMetrics = append(nodeMetrics, NewGaugeMetric(nodeValue, map[string]string{
				"resource_id":   node.Id,
				"resource_name": node.Name,
				"role":          node.Role,
				"status":        node.Status,
				"instance_name": inst.Name,
			}))
		}

		volumeMetrics = append(volumeMetrics, NewGaugeMetric(float64(inst.Volume.Size), map[string]string{
			"resource_id":   inst.Id,
			"resource_name": inst.Name,
		}))
	}

	return []*dto.MetricFamily{
		NewGaugeMetricFamily("rds_instance_status", statusMetrics),
		NewGaugeMetricFamily("rds_node_status", nodeMetrics),
		NewGaugeMetricFamily("rds_volume_size_gb", volumeMetrics),
	}
}
