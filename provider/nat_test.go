package provider

import (
	"testing"

	natGW "github.com/opentelekomcloud/gophertelekomcloud/openstack/networking/v2/extensions/natgateways"
)

func TestConvertNATGatewaysToMetrics(t *testing.T) {
	gateways := []natGW.NatGateway{
		{ID: "nat-001", Name: "nat-prod", Status: "ACTIVE", Spec: "1"},
		{ID: "nat-002", Name: "nat-dev", Status: "INACTIVE", Spec: "2"},
	}

	families := convertNATGatewaysToMetrics(gateways)

	if len(families) != 1 {
		t.Fatalf("expected 1 family, got %d", len(families))
	}
	if families[0].GetName() != "nat_gateway_status" {
		t.Errorf("expected family name %q, got %q", "nat_gateway_status", families[0].GetName())
	}
	if len(families[0].Metric) != 2 {
		t.Fatalf("expected 2 metrics, got %d", len(families[0].Metric))
	}

	// Build a lookup by resource_id for stable assertions.
	byID := make(map[string]float64)
	specByID := make(map[string]string)
	statusLabelByID := make(map[string]string)
	for _, m := range families[0].Metric {
		var rid, spec, status string
		for _, lp := range m.Label {
			switch lp.GetName() {
			case "resource_id":
				rid = lp.GetValue()
			case "spec":
				spec = lp.GetValue()
			case "status":
				status = lp.GetValue()
			}
		}
		if rid != "" {
			byID[rid] = m.Gauge.GetValue()
			specByID[rid] = spec
			statusLabelByID[rid] = status
		}
	}

	if val, ok := byID["nat-001"]; !ok || val != 0.0 {
		t.Errorf("expected ACTIVE gateway nat-001 to have value 0.0, got %v (exists=%v)", val, ok)
	}
	if val, ok := byID["nat-002"]; !ok || val != 1.0 {
		t.Errorf("expected INACTIVE gateway nat-002 to have value 1.0, got %v (exists=%v)", val, ok)
	}
	if specByID["nat-001"] != "1" {
		t.Errorf("expected nat-001 spec %q, got %q", "1", specByID["nat-001"])
	}
	if specByID["nat-002"] != "2" {
		t.Errorf("expected nat-002 spec %q, got %q", "2", specByID["nat-002"])
	}
	if statusLabelByID["nat-001"] != "ACTIVE" {
		t.Errorf("expected nat-001 status label %q, got %q", "ACTIVE", statusLabelByID["nat-001"])
	}
	if statusLabelByID["nat-002"] != "INACTIVE" {
		t.Errorf("expected nat-002 status label %q, got %q", "INACTIVE", statusLabelByID["nat-002"])
	}
}

func TestBuildNATNameMap(t *testing.T) {
	gateways := []natGW.NatGateway{
		{ID: "nat-001", Name: "nat-prod"},
		{ID: "nat-002", Name: "nat-dev"},
	}

	nameMap := buildNATNameMap(gateways)

	if len(nameMap) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(nameMap))
	}
	if nameMap["nat-001"] != "nat-prod" {
		t.Errorf("expected nat-001 -> %q, got %q", "nat-prod", nameMap["nat-001"])
	}
	if nameMap["nat-002"] != "nat-dev" {
		t.Errorf("expected nat-002 -> %q, got %q", "nat-dev", nameMap["nat-002"])
	}
}
