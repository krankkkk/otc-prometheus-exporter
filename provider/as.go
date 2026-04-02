package provider

import (
	"context"
	"fmt"

	asGroups "github.com/opentelekomcloud/gophertelekomcloud/openstack/autoscaling/v1/groups"
	dto "github.com/prometheus/client_model/go"

	"github.com/iits-consulting/otc-prometheus-exporter/otcclient"
)

// ASProvider collects service-API metrics for the OTC Auto Scaling service.
// CES has no AS metrics, so this provider only reports group instance counts and status.
type ASProvider struct{}

func (p *ASProvider) Namespace() string { return "SYS.AS" }

func (p *ASProvider) Collect(ctx context.Context, client *otcclient.Client) ([]*dto.MetricFamily, error) {
	// This provider has no CES metrics — it only uses service APIs.
	// When ShouldEnrich is false (enrich=false query param), it returns no data.
	if !ShouldEnrich(ctx) {
		return nil, nil
	}

	asClient, err := client.AutoScalingV1()
	if err != nil {
		return nil, fmt.Errorf("as client: %w", err)
	}

	resp, err := asGroups.List(asClient, asGroups.ListOpts{})
	if err != nil {
		return nil, fmt.Errorf("listing AS groups: %w", err)
	}

	client.Logger.Debug("service API completed", "namespace", "SYS.AS", "groups", len(resp.ScalingGroups))

	return convertASGroupsToMetrics(resp.ScalingGroups), nil
}

// convertASGroupsToMetrics creates MetricFamily objects for Auto Scaling group metrics:
// - as_group_actual_instances: current number of instances
// - as_group_desired_instances: desired number of instances
// - as_group_min_instances: minimum number of instances
// - as_group_max_instances: maximum number of instances
// - as_group_status: 0.0 if INSERVICE, 1.0 otherwise (OTC convention: 0=normal, 1=abnormal)
func convertASGroupsToMetrics(groups []asGroups.Group) []*dto.MetricFamily {
	actualMetrics := make([]*dto.Metric, 0, len(groups))
	desiredMetrics := make([]*dto.Metric, 0, len(groups))
	minMetrics := make([]*dto.Metric, 0, len(groups))
	maxMetrics := make([]*dto.Metric, 0, len(groups))
	statusMetrics := make([]*dto.Metric, 0, len(groups))

	for _, g := range groups {
		labels := map[string]string{
			"resource_id":   g.ID,
			"resource_name": g.Name,
		}

		actualMetrics = append(actualMetrics, NewGaugeMetric(float64(g.ActualInstanceNumber), labels))
		desiredMetrics = append(desiredMetrics, NewGaugeMetric(float64(g.DesireInstanceNumber), labels))
		minMetrics = append(minMetrics, NewGaugeMetric(float64(g.MinInstanceNumber), labels))
		maxMetrics = append(maxMetrics, NewGaugeMetric(float64(g.MaxInstanceNumber), labels))

		statusValue := 1.0
		if g.Status == "INSERVICE" {
			statusValue = 0.0
		}
		statusMetrics = append(statusMetrics, NewGaugeMetric(statusValue, map[string]string{
			"resource_id":   g.ID,
			"resource_name": g.Name,
			"status":        g.Status,
		}))
	}

	return []*dto.MetricFamily{
		NewGaugeMetricFamily("as_group_actual_instances", actualMetrics),
		NewGaugeMetricFamily("as_group_desired_instances", desiredMetrics),
		NewGaugeMetricFamily("as_group_min_instances", minMetrics),
		NewGaugeMetricFamily("as_group_max_instances", maxMetrics),
		NewGaugeMetricFamily("as_group_status", statusMetrics),
	}
}

