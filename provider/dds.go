package provider

import (
	"context"

	ddsInst "github.com/opentelekomcloud/gophertelekomcloud/openstack/dds/v3/instances"
	dto "github.com/prometheus/client_model/go"

	"github.com/iits-consulting/otc-prometheus-exporter/otcclient"
)

// DDSProvider collects CES metrics for the OTC Document Database Service,
// enriches them with instance/group/node names, and reports instance and node status.
type DDSProvider struct{}

func (p *DDSProvider) Namespace() string { return "SYS.DDS" }

func (p *DDSProvider) Collect(ctx context.Context, client *otcclient.Client) ([]*dto.MetricFamily, error) {
	return CollectWithEnrichment(ctx, client, "SYS.DDS", func(ctx context.Context, client *otcclient.Client) (*EnrichResult, error) {
		ddsClient, err := client.DDSV3()
		if err != nil {
			return nil, err
		}
		resp, err := ddsInst.List(ddsClient, ddsInst.ListInstanceOpts{})
		if err != nil {
			return nil, err
		}
		return &EnrichResult{
			NameMap:       buildDDSNameMap(resp.Instances),
			ExtraFamilies: convertDDSInstancesToMetrics(resp.Instances),
		}, nil
	})
}

// buildDDSNameMap creates a mapping from DDS instance, group, and node IDs to names.
func buildDDSNameMap(instances []ddsInst.InstanceResponse) map[string]string {
	m := make(map[string]string)
	for _, inst := range instances {
		m[inst.Id] = inst.Name
		for _, g := range inst.Groups {
			if g.Name != "" {
				m[g.Id] = g.Name
			}
			for _, n := range g.Nodes {
				if n.Name != "" {
					m[n.Id] = n.Name
				}
			}
		}
	}
	return m
}

// convertDDSInstancesToMetrics creates MetricFamily objects for DDS-specific metrics:
// - dds_instance_status: 0.0 if normal, 1.0 otherwise (OTC convention: 0=normal, 1=abnormal)
// - dds_node_status: per-node status with role information
func convertDDSInstancesToMetrics(instances []ddsInst.InstanceResponse) []*dto.MetricFamily {
	statusMetrics := make([]*dto.Metric, 0, len(instances))
	nodeMetrics := make([]*dto.Metric, 0)

	for _, inst := range instances {
		statusValue := 1.0
		if inst.Status == "normal" {
			statusValue = 0.0
		}
		statusMetrics = append(statusMetrics, NewGaugeMetric(statusValue, map[string]string{
			"resource_id":   inst.Id,
			"resource_name": inst.Name,
			"status":        inst.Status,
		}))

		for _, g := range inst.Groups {
			for _, n := range g.Nodes {
				nodeValue := 1.0
				if n.Status == "normal" {
					nodeValue = 0.0
				}
				nodeMetrics = append(nodeMetrics, NewGaugeMetric(nodeValue, map[string]string{
					"resource_id":   n.Id,
					"resource_name": n.Name,
					"role":          n.Role,
					"status":        n.Status,
					"instance_name": inst.Name,
				}))
			}
		}
	}

	return []*dto.MetricFamily{
		NewGaugeMetricFamily("dds_instance_status", statusMetrics),
		NewGaugeMetricFamily("dds_node_status", nodeMetrics),
	}
}

func (p *DDSProvider) Dashboard() DashboardConfig {
	return DashboardConfig{
		Title: "DDS - Document Database Service",
		UID:   "otc-dds",
		Sections: []PanelSection{
			{Title: "Overview", Panels: []PanelConfig{
				{Metric: "dds_instance_status", Title: "Instance Status", Unit: "short", Type: Stat,
					Thresholds: []Threshold{{Value: 0, Color: "green"}, {Value: 1, Color: "red"}}},
			}},
			{Title: "Node Status", Panels: []PanelConfig{
				{Metric: "dds_node_status", Title: "Node Status", Unit: "short", Type: Table,
					Thresholds: []Threshold{{Value: 0, Color: "green"}, {Value: 1, Color: "red"}}},
			}},
		},
	}
}
