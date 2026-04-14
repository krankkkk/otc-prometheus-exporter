package provider

import (
	"testing"

	ddsInst "github.com/opentelekomcloud/gophertelekomcloud/openstack/dds/v3/instances"
)

func TestConvertDDSInstancesToMetrics(t *testing.T) {
	instances := []ddsInst.InstanceResponse{
		{
			Id:     "dds-001",
			Name:   "mongo-prod",
			Status: "normal",
			Groups: []ddsInst.Group{
				{
					Id:   "grp-001",
					Name: "replica",
					Nodes: []ddsInst.Nodes{
						{Id: "node-001", Name: "mongo-prod-node-1", Role: "Primary", Status: "normal"},
						{Id: "node-002", Name: "mongo-prod-node-2", Role: "Secondary", Status: "normal"},
					},
				},
			},
		},
		{
			Id:     "dds-002",
			Name:   "mongo-dev",
			Status: "abnormal",
			Groups: []ddsInst.Group{
				{
					Id:   "grp-002",
					Name: "replica",
					Nodes: []ddsInst.Nodes{
						{Id: "node-003", Name: "mongo-dev-node-1", Role: "Primary", Status: "abnormal"},
					},
				},
			},
		},
	}

	families := convertDDSInstancesToMetrics(instances)

	if len(families) != 2 {
		t.Fatalf("expected 2 families, got %d", len(families))
	}

	// Verify family names.
	if families[0].GetName() != "dds_instance_status" {
		t.Errorf("expected family[0] name %q, got %q", "dds_instance_status", families[0].GetName())
	}
	if families[1].GetName() != "dds_node_status" {
		t.Errorf("expected family[1] name %q, got %q", "dds_node_status", families[1].GetName())
	}

	// dds_instance_status: 2 metrics (one per instance).
	if len(families[0].Metric) != 2 {
		t.Fatalf("dds_instance_status: expected 2 metrics, got %d", len(families[0].Metric))
	}

	// dds_node_status: 3 metrics total (2 nodes in dds-001 + 1 node in dds-002).
	if len(families[1].Metric) != 3 {
		t.Fatalf("dds_node_status: expected 3 metrics, got %d", len(families[1].Metric))
	}

	// Build a lookup by resource_id for instance status assertions.
	instanceStatusByID := make(map[string]float64)
	for _, m := range families[0].Metric {
		for _, lp := range m.Label {
			if lp.GetName() == "resource_id" {
				instanceStatusByID[lp.GetValue()] = m.Gauge.GetValue()
			}
		}
	}

	if val, ok := instanceStatusByID["dds-001"]; !ok || val != 0.0 {
		t.Errorf("expected normal instance dds-001 to have value 0.0, got %v (exists=%v)", val, ok)
	}
	if val, ok := instanceStatusByID["dds-002"]; !ok || val != 1.0 {
		t.Errorf("expected abnormal instance dds-002 to have value 1.0, got %v (exists=%v)", val, ok)
	}

	// Build a lookup by resource_id for node status assertions.
	nodeStatusByID := make(map[string]float64)
	nodeRoleByID := make(map[string]string)
	nodeInstanceByID := make(map[string]string)
	for _, m := range families[1].Metric {
		var rid, role, instName string
		for _, lp := range m.Label {
			switch lp.GetName() {
			case "resource_id":
				rid = lp.GetValue()
			case "role":
				role = lp.GetValue()
			case "instance_name":
				instName = lp.GetValue()
			}
		}
		if rid != "" {
			nodeStatusByID[rid] = m.Gauge.GetValue()
			nodeRoleByID[rid] = role
			nodeInstanceByID[rid] = instName
		}
	}

	if val, ok := nodeStatusByID["node-001"]; !ok || val != 0.0 {
		t.Errorf("expected normal node node-001 to have value 0.0, got %v (exists=%v)", val, ok)
	}
	if val, ok := nodeStatusByID["node-003"]; !ok || val != 1.0 {
		t.Errorf("expected abnormal node node-003 to have value 1.0, got %v (exists=%v)", val, ok)
	}
	if nodeRoleByID["node-001"] != "Primary" {
		t.Errorf("expected node-001 role %q, got %q", "Primary", nodeRoleByID["node-001"])
	}
	if nodeRoleByID["node-002"] != "Secondary" {
		t.Errorf("expected node-002 role %q, got %q", "Secondary", nodeRoleByID["node-002"])
	}
	if nodeInstanceByID["node-001"] != "mongo-prod" {
		t.Errorf("expected node-001 instance_name %q, got %q", "mongo-prod", nodeInstanceByID["node-001"])
	}
	if nodeInstanceByID["node-003"] != "mongo-dev" {
		t.Errorf("expected node-003 instance_name %q, got %q", "mongo-dev", nodeInstanceByID["node-003"])
	}
}

func TestBuildDDSNameMap(t *testing.T) {
	instances := []ddsInst.InstanceResponse{
		{
			Id:   "dds-001",
			Name: "mongo-prod",
			Groups: []ddsInst.Group{
				{
					Id:   "grp-001",
					Name: "replica",
					Nodes: []ddsInst.Nodes{
						{Id: "node-001", Name: "mongo-prod-node-1"},
						{Id: "node-002", Name: "mongo-prod-node-2"},
					},
				},
			},
		},
		{
			Id:   "dds-002",
			Name: "mongo-dev",
			Groups: []ddsInst.Group{
				{
					Id:   "grp-002",
					Name: "", // empty group name should not be added
					Nodes: []ddsInst.Nodes{
						{Id: "node-003", Name: "mongo-dev-node-1"},
						{Id: "node-004", Name: ""}, // empty node name should not be added
					},
				},
			},
		},
	}

	nameMap := buildDDSNameMap(instances)

	// Expected: dds-001, grp-001, node-001, node-002, dds-002, node-003 = 6 entries.
	// grp-002 is excluded (empty name), node-004 is excluded (empty name).
	if len(nameMap) != 6 {
		t.Fatalf("expected 6 entries, got %d: %v", len(nameMap), nameMap)
	}

	checks := map[string]string{
		"dds-001":  "mongo-prod",
		"dds-002":  "mongo-dev",
		"grp-001":  "replica",
		"node-001": "mongo-prod-node-1",
		"node-002": "mongo-prod-node-2",
		"node-003": "mongo-dev-node-1",
	}
	for id, expectedName := range checks {
		if nameMap[id] != expectedName {
			t.Errorf("expected %s -> %q, got %q", id, expectedName, nameMap[id])
		}
	}

	// Confirm empty-name entries were not added.
	if _, ok := nameMap["grp-002"]; ok {
		t.Errorf("expected grp-002 (empty name) to be absent from nameMap")
	}
	if _, ok := nameMap["node-004"]; ok {
		t.Errorf("expected node-004 (empty name) to be absent from nameMap")
	}
}
