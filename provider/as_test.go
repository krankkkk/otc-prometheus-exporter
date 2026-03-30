package provider

import (
	"testing"

	asGroups "github.com/opentelekomcloud/gophertelekomcloud/openstack/autoscaling/v1/groups"
)

func TestConvertASGroupsToMetrics(t *testing.T) {
	groups := []asGroups.Group{
		{
			ID:                   "as-001",
			Name:                 "web-scaling-group",
			Status:               "INSERVICE",
			ActualInstanceNumber: 3,
			DesireInstanceNumber: 3,
			MinInstanceNumber:    1,
			MaxInstanceNumber:    10,
		},
		{
			ID:                   "as-002",
			Name:                 "worker-scaling-group",
			Status:               "PAUSED",
			ActualInstanceNumber: 0,
			DesireInstanceNumber: 2,
			MinInstanceNumber:    0,
			MaxInstanceNumber:    5,
		},
	}

	families := convertASGroupsToMetrics(groups)

	if len(families) != 5 {
		t.Fatalf("expected 5 families, got %d", len(families))
	}

	// Verify family names.
	expectedNames := []string{
		"as_group_actual_instances",
		"as_group_desired_instances",
		"as_group_min_instances",
		"as_group_max_instances",
		"as_group_status",
	}
	for i, name := range expectedNames {
		if families[i].GetName() != name {
			t.Errorf("expected family[%d] name %q, got %q", i, name, families[i].GetName())
		}
	}

	// Each family should have 2 metrics (one per group).
	for i, fam := range families {
		if len(fam.Metric) != 2 {
			t.Errorf("family[%d] %q: expected 2 metrics, got %d", i, fam.GetName(), len(fam.Metric))
		}
	}

	// Verify actual instance counts for first group.
	if families[0].Metric[0].Gauge.GetValue() != 3.0 {
		t.Errorf("expected actual instances 3.0, got %f", families[0].Metric[0].Gauge.GetValue())
	}

	// Verify status values: INSERVICE=0.0, PAUSED=1.0 (OTC convention: 0=normal, 1=abnormal).
	statusFam := families[4]
	byID := make(map[string]float64)
	for _, m := range statusFam.Metric {
		for _, lp := range m.Label {
			if lp.GetName() == "resource_id" {
				byID[lp.GetValue()] = m.Gauge.GetValue()
			}
		}
	}
	if val, ok := byID["as-001"]; !ok || val != 0.0 {
		t.Errorf("expected INSERVICE group as-001 to have value 0.0, got %v (exists=%v)", val, ok)
	}
	if val, ok := byID["as-002"]; !ok || val != 1.0 {
		t.Errorf("expected PAUSED group as-002 to have value 1.0, got %v (exists=%v)", val, ok)
	}
}
