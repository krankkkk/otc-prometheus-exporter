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

func (p *RDSProvider) Dashboard() DashboardConfig {
	return DashboardConfig{
		Title: "RDS - Relational Database Service",
		UID:   "otc-rds",
		Sections: []PanelSection{
			{Title: "Overview", Panels: []PanelConfig{
				{Metric: "rds_instance_status", Title: "Instance Status", Unit: "short", Type: Stat,
					Thresholds: []Threshold{{Value: 0, Color: "green"}, {Value: 1, Color: "red"}}},
				{Metric: "rds_volume_size_gb", Title: "Volume Size", Unit: "decgbytes", Type: Gauge},
			}},
			{Title: "Node Status", Panels: []PanelConfig{
				{Metric: "rds_node_status", Title: "Node Status", Unit: "short", Type: Table,
					Thresholds: []Threshold{{Value: 0, Color: "green"}, {Value: 1, Color: "red"}}},
			}},
			{Title: "Performance", Panels: []PanelConfig{
				{Metric: "rds_rds001_cpu_util", Title: "CPU Utilization", Unit: "percent", Type: TimeSeries,
					Thresholds: []Threshold{{Value: 0, Color: "green"}, {Value: 80, Color: "yellow"}, {Value: 90, Color: "red"}}},
				{Metric: "rds_rds002_mem_util", Title: "Memory Utilization", Unit: "percent", Type: TimeSeries,
					Thresholds: []Threshold{{Value: 0, Color: "green"}, {Value: 80, Color: "yellow"}, {Value: 90, Color: "red"}}},
				{Metric: "rds_rds039_disk_util", Title: "Disk Utilization", Unit: "percent", Type: TimeSeries,
					Thresholds: []Threshold{{Value: 0, Color: "green"}, {Value: 80, Color: "yellow"}, {Value: 90, Color: "red"}}},
			}},
			{Title: "Connections", Panels: []PanelConfig{
				{Metric: "rds_rds042_database_connections", Title: "Active Connections", Unit: "short", Type: TimeSeries},
				{Metric: "rds_rds083_conn_usage", Title: "Connection Usage", Unit: "percent", Type: TimeSeries,
					Thresholds: []Threshold{{Value: 0, Color: "green"}, {Value: 80, Color: "yellow"}, {Value: 90, Color: "red"}}},
			}},
			{Title: "Replication", Panels: []PanelConfig{
				{Metric: "rds_rds046_replication_lag", Title: "Replication Lag", Unit: "s", Type: TimeSeries,
					Thresholds: []Threshold{{Value: 0, Color: "green"}, {Value: 10, Color: "yellow"}, {Value: 60, Color: "red"}}},
			}},
			{Title: "I/O", Panels: []PanelConfig{
				{Metric: "rds_rds003_iops", Title: "IOPS", Unit: "iops", Type: TimeSeries},
				{Metric: "rds_rds004_bytes_in", Title: "Bytes In", Unit: "Bps", Type: TimeSeries},
				{Metric: "rds_rds005_bytes_out", Title: "Bytes Out", Unit: "Bps", Type: TimeSeries},
			}},
		},
	}
}
