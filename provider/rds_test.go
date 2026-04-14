package provider

import (
	"testing"

	rdsInstances "github.com/opentelekomcloud/gophertelekomcloud/openstack/rds/v3/instances"
	dto "github.com/prometheus/client_model/go"
)

func TestConvertRDSInstancesToMetrics(t *testing.T) {
	instances := []rdsInstances.InstanceResponse{
		{
			Id:     "rds-001",
			Name:   "my-mysql",
			Status: "ACTIVE",
			Nodes: []rdsInstances.Nodes{
				{Id: "node-1", Name: "my-mysql-node-1", Role: "master", Status: "ACTIVE"},
				{Id: "node-2", Name: "my-mysql-node-2", Role: "slave", Status: "ACTIVE"},
			},
			Volume: rdsInstances.Volume{Size: 100},
		},
		{
			Id:     "rds-002",
			Name:   "my-postgres",
			Status: "FAILED",
			Nodes: []rdsInstances.Nodes{
				{Id: "node-3", Name: "my-postgres-node-1", Role: "master", Status: "FAILED"},
			},
			Volume: rdsInstances.Volume{Size: 200},
		},
	}

	families := convertRDSInstancesToMetrics(instances)

	if len(families) != 3 {
		t.Fatalf("expected 3 families, got %d", len(families))
	}

	// Verify family names.
	expectedNames := []string{"rds_instance_status", "rds_node_status", "rds_volume_size_gb"}
	for i, name := range expectedNames {
		if families[i].GetName() != name {
			t.Errorf("expected family[%d] name %q, got %q", i, name, families[i].GetName())
		}
	}

	// rds_instance_status: 2 instances.
	if len(families[0].Metric) != 2 {
		t.Errorf("expected 2 instance status metrics, got %d", len(families[0].Metric))
	}

	// rds_node_status: 3 nodes total (2 + 1).
	if len(families[1].Metric) != 3 {
		t.Errorf("expected 3 node status metrics, got %d", len(families[1].Metric))
	}

	// rds_volume_size_gb: 2 instances.
	if len(families[2].Metric) != 2 {
		t.Errorf("expected 2 volume metrics, got %d", len(families[2].Metric))
	}

	// Verify ACTIVE instance has value 0.0 (OTC convention: 0=normal).
	activeMetric := families[0].Metric[0]
	if activeMetric.Gauge.GetValue() != 0.0 {
		t.Errorf("expected ACTIVE instance value 0.0, got %f", activeMetric.Gauge.GetValue())
	}

	// Verify FAILED instance has value 1.0 (OTC convention: 1=abnormal).
	failedMetric := families[0].Metric[1]
	if failedMetric.Gauge.GetValue() != 1.0 {
		t.Errorf("expected FAILED instance value 1.0, got %f", failedMetric.Gauge.GetValue())
	}

	// Verify rds_node_status labels use node-level identifiers, not instance-level.
	getLabelValue := func(m *dto.Metric, name string) string {
		for _, lp := range m.Label {
			if lp.GetName() == name {
				return lp.GetValue()
			}
		}
		return ""
	}

	// First node metric: node-1 of my-mysql.
	node0 := families[1].Metric[0]
	if got := getLabelValue(node0, "resource_id"); got != "node-1" {
		t.Errorf("node0 resource_id: expected %q, got %q", "node-1", got)
	}
	if got := getLabelValue(node0, "resource_name"); got != "my-mysql-node-1" {
		t.Errorf("node0 resource_name: expected %q, got %q", "my-mysql-node-1", got)
	}
	if got := getLabelValue(node0, "instance_name"); got != "my-mysql" {
		t.Errorf("node0 instance_name: expected %q, got %q", "my-mysql", got)
	}

	// Third node metric: node-3 of my-postgres (FAILED -> value 0.0).
	node2 := families[1].Metric[2]
	if got := getLabelValue(node2, "resource_id"); got != "node-3" {
		t.Errorf("node2 resource_id: expected %q, got %q", "node-3", got)
	}
	if got := getLabelValue(node2, "resource_name"); got != "my-postgres-node-1" {
		t.Errorf("node2 resource_name: expected %q, got %q", "my-postgres-node-1", got)
	}
	if got := getLabelValue(node2, "instance_name"); got != "my-postgres" {
		t.Errorf("node2 instance_name: expected %q, got %q", "my-postgres", got)
	}
	if node2.Gauge.GetValue() != 1.0 {
		t.Errorf("expected FAILED node value 1.0, got %f", node2.Gauge.GetValue())
	}
}

func TestBuildRDSNameMap(t *testing.T) {
	instances := []rdsInstances.InstanceResponse{
		{Id: "rds-001", Name: "my-mysql"},
		{Id: "rds-002", Name: "my-postgres"},
	}

	nameMap := buildRDSNameMap(instances)

	if len(nameMap) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(nameMap))
	}
	if nameMap["rds-001"] != "my-mysql" {
		t.Errorf("expected rds-001 -> %q, got %q", "my-mysql", nameMap["rds-001"])
	}
	if nameMap["rds-002"] != "my-postgres" {
		t.Errorf("expected rds-002 -> %q, got %q", "my-postgres", nameMap["rds-002"])
	}
}

func TestBuildRDSNameMapNodeLevel(t *testing.T) {
	instances := []rdsInstances.InstanceResponse{
		{
			Id:   "rds-001",
			Name: "my-mysql",
			Nodes: []rdsInstances.Nodes{
				{Id: "node-a1", Name: "node-primary"},
				{Id: "node-a2", Name: "node-replica"},
			},
		},
	}

	nameMap := buildRDSNameMap(instances)

	// 1 instance + 2 nodes = 3 entries.
	if len(nameMap) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(nameMap))
	}

	// Instance-level mapping.
	if nameMap["rds-001"] != "my-mysql" {
		t.Errorf("expected rds-001 -> %q, got %q", "my-mysql", nameMap["rds-001"])
	}

	// Node-level mappings use "instance-name/node-name" format.
	if nameMap["node-a1"] != "my-mysql/node-primary" {
		t.Errorf("expected node-a1 -> %q, got %q", "my-mysql/node-primary", nameMap["node-a1"])
	}
	if nameMap["node-a2"] != "my-mysql/node-replica" {
		t.Errorf("expected node-a2 -> %q, got %q", "my-mysql/node-replica", nameMap["node-a2"])
	}
}

func TestRDSCacheIntegration(t *testing.T) {
	cache := NewNameCache()
	names := buildRDSNameMap([]rdsInstances.InstanceResponse{
		{
			Id:   "rds-001",
			Name: "my-mysql",
			Nodes: []rdsInstances.Nodes{
				{Id: "node-a1", Name: "node-primary"},
			},
		},
	})
	cache.Put("SYS.RDS", names)

	got := cache.Get("SYS.RDS")
	if len(got) != 2 {
		t.Fatalf("expected 2 cached entries (1 instance + 1 node), got %d", len(got))
	}
	if got["rds-001"] != "my-mysql" {
		t.Errorf("expected rds-001 -> my-mysql, got %q", got["rds-001"])
	}
	if got["node-a1"] != "my-mysql/node-primary" {
		t.Errorf("expected node-a1 -> my-mysql/node-primary, got %q", got["node-a1"])
	}
}
