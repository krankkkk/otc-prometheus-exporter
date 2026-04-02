package provider

import (
	"context"

	dcsLifecycle "github.com/opentelekomcloud/gophertelekomcloud/openstack/dcs/v1/lifecycle"
	dto "github.com/prometheus/client_model/go"

	"github.com/iits-consulting/otc-prometheus-exporter/otcclient"
)

// DCSProvider collects CES metrics for the OTC Distributed Cache Service,
// enriches them with instance names, and reports instance status and capacity.
type DCSProvider struct{}

func (p *DCSProvider) Namespace() string { return "SYS.DCS" }

func (p *DCSProvider) Collect(ctx context.Context, client *otcclient.Client) ([]*dto.MetricFamily, error) {
	return CollectWithEnrichment(ctx, client, "SYS.DCS", func(ctx context.Context, client *otcclient.Client) (*EnrichResult, error) {
		dcsClient, err := client.DCSV1()
		if err != nil {
			return nil, err
		}
		resp, err := dcsLifecycle.List(dcsClient, dcsLifecycle.ListDcsInstanceOpts{})
		if err != nil {
			return nil, err
		}
		return &EnrichResult{
			NameMap:       buildDCSNameMap(resp.Instances),
			ExtraFamilies: convertDCSInstancesToMetrics(resp.Instances),
		}, nil
	})
}

// buildDCSNameMap creates a mapping from DCS instance ID to instance name.
func buildDCSNameMap(instances []dcsLifecycle.Instance) map[string]string {
	m := make(map[string]string, len(instances))
	for _, inst := range instances {
		m[inst.InstanceID] = inst.Name
	}
	return m
}

// convertDCSInstancesToMetrics creates MetricFamily objects for DCS-specific metrics:
// - dcs_instance_status: 0.0 if RUNNING, 1.0 otherwise (OTC convention: 0=normal, 1=abnormal)
// - dcs_instance_capacity_mb: cache capacity in MB
func convertDCSInstancesToMetrics(instances []dcsLifecycle.Instance) []*dto.MetricFamily {
	statusMetrics := make([]*dto.Metric, 0, len(instances))
	capacityMetrics := make([]*dto.Metric, 0, len(instances))

	for _, inst := range instances {
		statusValue := 1.0
		if inst.Status == "RUNNING" {
			statusValue = 0.0
		}
		statusMetrics = append(statusMetrics, NewGaugeMetric(statusValue, map[string]string{
			"resource_id":   inst.InstanceID,
			"resource_name": inst.Name,
			"status":        inst.Status,
		}))

		capacityMetrics = append(capacityMetrics, NewGaugeMetric(float64(inst.Capacity), map[string]string{
			"resource_id":   inst.InstanceID,
			"resource_name": inst.Name,
		}))
	}

	return []*dto.MetricFamily{
		NewGaugeMetricFamily("dcs_instance_status", statusMetrics),
		NewGaugeMetricFamily("dcs_instance_capacity_mb", capacityMetrics),
	}
}

