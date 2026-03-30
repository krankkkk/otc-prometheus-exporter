package provider

import (
	"testing"

	vpcBW "github.com/opentelekomcloud/gophertelekomcloud/openstack/vpc/v1/bandwidths"
)

func TestConvertVPCBandwidthsToMetrics(t *testing.T) {
	bandwidths := []vpcBW.BandWidth{
		{ID: "bw-001", Name: "prod-bandwidth", Size: 100},
		{ID: "bw-002", Name: "dev-bandwidth", Size: 10},
	}

	families := convertVPCBandwidthsToMetrics(bandwidths)

	if len(families) != 1 {
		t.Fatalf("expected 1 family, got %d", len(families))
	}
	if families[0].GetName() != "vpc_bandwidth_size_mbit" {
		t.Errorf("expected family name %q, got %q", "vpc_bandwidth_size_mbit", families[0].GetName())
	}
	if len(families[0].Metric) != 2 {
		t.Fatalf("expected 2 metrics, got %d", len(families[0].Metric))
	}

	// Build a lookup by resource_id for stable assertions.
	byID := make(map[string]float64)
	for _, m := range families[0].Metric {
		for _, lp := range m.Label {
			if lp.GetName() == "resource_id" {
				byID[lp.GetValue()] = m.Gauge.GetValue()
			}
		}
	}

	if val, ok := byID["bw-001"]; !ok || val != 100.0 {
		t.Errorf("expected bw-001 size 100.0, got %v (exists=%v)", val, ok)
	}
	if val, ok := byID["bw-002"]; !ok || val != 10.0 {
		t.Errorf("expected bw-002 size 10.0, got %v (exists=%v)", val, ok)
	}

	// Verify resource_name label is present.
	resourceNameByID := make(map[string]string)
	for _, m := range families[0].Metric {
		var rid, rname string
		for _, lp := range m.Label {
			switch lp.GetName() {
			case "resource_id":
				rid = lp.GetValue()
			case "resource_name":
				rname = lp.GetValue()
			}
		}
		if rid != "" {
			resourceNameByID[rid] = rname
		}
	}

	if resourceNameByID["bw-001"] != "prod-bandwidth" {
		t.Errorf("expected bw-001 resource_name %q, got %q", "prod-bandwidth", resourceNameByID["bw-001"])
	}
	if resourceNameByID["bw-002"] != "dev-bandwidth" {
		t.Errorf("expected bw-002 resource_name %q, got %q", "dev-bandwidth", resourceNameByID["bw-002"])
	}
}

func TestBuildVPCNameMap(t *testing.T) {
	bandwidths := []vpcBW.BandWidth{
		{ID: "bw-001", Name: "prod-bandwidth"},
		{ID: "bw-002", Name: "dev-bandwidth"},
	}

	nameMap := buildVPCNameMap(bandwidths)

	if len(nameMap) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(nameMap))
	}
	if nameMap["bw-001"] != "prod-bandwidth" {
		t.Errorf("expected bw-001 -> %q, got %q", "prod-bandwidth", nameMap["bw-001"])
	}
	if nameMap["bw-002"] != "dev-bandwidth" {
		t.Errorf("expected bw-002 -> %q, got %q", "dev-bandwidth", nameMap["bw-002"])
	}
}
