package provider

import (
	"testing"

	dcsLifecycle "github.com/opentelekomcloud/gophertelekomcloud/openstack/dcs/v1/lifecycle"
)

func TestConvertDCSInstancesToMetrics(t *testing.T) {
	instances := []dcsLifecycle.Instance{
		{InstanceID: "dcs-001", Name: "redis-prod", Status: "RUNNING", Capacity: 4},
		{InstanceID: "dcs-002", Name: "redis-dev", Status: "CREATING", Capacity: 2},
	}

	families := convertDCSInstancesToMetrics(instances)

	if len(families) != 2 {
		t.Fatalf("expected 2 families, got %d", len(families))
	}

	// Verify family names.
	expectedNames := []string{"dcs_instance_status", "dcs_instance_capacity_mb"}
	for i, name := range expectedNames {
		if families[i].GetName() != name {
			t.Errorf("expected family[%d] name %q, got %q", i, name, families[i].GetName())
		}
	}

	// Each family should have 2 metrics (one per instance).
	for i, fam := range families {
		if len(fam.Metric) != 2 {
			t.Errorf("family[%d] %q: expected 2 metrics, got %d", i, fam.GetName(), len(fam.Metric))
		}
	}

	// Build a lookup by resource_id for stable assertions on status family.
	statusByID := make(map[string]float64)
	for _, m := range families[0].Metric {
		for _, lp := range m.Label {
			if lp.GetName() == "resource_id" {
				statusByID[lp.GetValue()] = m.Gauge.GetValue()
			}
		}
	}

	if val, ok := statusByID["dcs-001"]; !ok || val != 0.0 {
		t.Errorf("expected RUNNING instance dcs-001 to have status value 0.0, got %v (exists=%v)", val, ok)
	}
	if val, ok := statusByID["dcs-002"]; !ok || val != 1.0 {
		t.Errorf("expected CREATING instance dcs-002 to have status value 1.0, got %v (exists=%v)", val, ok)
	}

	// Verify status label is present.
	for _, m := range families[0].Metric {
		found := false
		for _, lp := range m.Label {
			if lp.GetName() == "status" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected status label on dcs_instance_status metric")
		}
	}

	// Build a lookup by resource_id for capacity family.
	capacityByID := make(map[string]float64)
	for _, m := range families[1].Metric {
		for _, lp := range m.Label {
			if lp.GetName() == "resource_id" {
				capacityByID[lp.GetValue()] = m.Gauge.GetValue()
			}
		}
	}

	if val, ok := capacityByID["dcs-001"]; !ok || val != 4.0 {
		t.Errorf("expected capacity 4.0 for dcs-001, got %v (exists=%v)", val, ok)
	}
	if val, ok := capacityByID["dcs-002"]; !ok || val != 2.0 {
		t.Errorf("expected capacity 2.0 for dcs-002, got %v (exists=%v)", val, ok)
	}
}

func TestBuildDCSNameMap(t *testing.T) {
	instances := []dcsLifecycle.Instance{
		{InstanceID: "dcs-001", Name: "redis-prod"},
		{InstanceID: "dcs-002", Name: "redis-dev"},
	}

	nameMap := buildDCSNameMap(instances)

	if len(nameMap) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(nameMap))
	}
	if nameMap["dcs-001"] != "redis-prod" {
		t.Errorf("expected dcs-001 -> %q, got %q", "redis-prod", nameMap["dcs-001"])
	}
	if nameMap["dcs-002"] != "redis-dev" {
		t.Errorf("expected dcs-002 -> %q, got %q", "redis-dev", nameMap["dcs-002"])
	}
}
