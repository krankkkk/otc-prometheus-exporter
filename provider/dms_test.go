package provider

import (
	"testing"

	dmsInstances "github.com/opentelekomcloud/gophertelekomcloud/openstack/dms/v2/instances/lifecycle"
)

func TestConvertDMSInstancesToMetrics(t *testing.T) {
	instances := []dmsInstances.Instance{
		{
			InstanceID:        "dms-001",
			Name:              "kafka-prod",
			UsedStorageSpace:  50,
			TotalStorageSpace: 200,
			PartitionNum:      "300",
		},
		{
			InstanceID:        "dms-002",
			Name:              "kafka-dev",
			UsedStorageSpace:  10,
			TotalStorageSpace: 100,
			PartitionNum:      "100",
		},
	}

	families := convertDMSInstancesToMetrics(instances)

	if len(families) != 3 {
		t.Fatalf("expected 3 families, got %d", len(families))
	}

	// Verify family names.
	expectedNames := []string{"dms_instance_storage_used_gb", "dms_instance_storage_total_gb", "dms_instance_partitions"}
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

	// Verify values for first instance in each family.
	if families[0].Metric[0].Gauge.GetValue() != 50.0 {
		t.Errorf("expected used storage 50.0, got %f", families[0].Metric[0].Gauge.GetValue())
	}
	if families[1].Metric[0].Gauge.GetValue() != 200.0 {
		t.Errorf("expected total storage 200.0, got %f", families[1].Metric[0].Gauge.GetValue())
	}
	if families[2].Metric[0].Gauge.GetValue() != 300.0 {
		t.Errorf("expected partitions 300.0, got %f", families[2].Metric[0].Gauge.GetValue())
	}
}

func TestBuildDMSNameMap(t *testing.T) {
	instances := []dmsInstances.Instance{
		{InstanceID: "dms-001", Name: "kafka-prod"},
		{InstanceID: "dms-002", Name: "kafka-dev"},
	}

	nameMap := buildDMSNameMap(instances)

	if len(nameMap) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(nameMap))
	}
	if nameMap["dms-001"] != "kafka-prod" {
		t.Errorf("expected dms-001 -> %q, got %q", "kafka-prod", nameMap["dms-001"])
	}
	if nameMap["dms-002"] != "kafka-dev" {
		t.Errorf("expected dms-002 -> %q, got %q", "kafka-dev", nameMap["dms-002"])
	}
}
