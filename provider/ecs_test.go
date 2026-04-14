package provider

import (
	"testing"

	otcCompute "github.com/opentelekomcloud/gophertelekomcloud/openstack/compute/v2/servers"
	dto "github.com/prometheus/client_model/go"
)

func TestConvertECSInstancesToMetrics(t *testing.T) {
	servers := []otcCompute.Server{
		{ID: "id-001", Name: "web-server", Status: "ACTIVE"},
		{ID: "id-002", Name: "db-server", Status: "SHUTOFF"},
	}

	families := convertECSInstancesToMetrics(servers)

	if len(families) != 1 {
		t.Fatalf("expected 1 family, got %d", len(families))
	}
	if families[0].GetName() != "ecs_instance_status" {
		t.Errorf("expected family name %q, got %q", "ecs_instance_status", families[0].GetName())
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

	if val, ok := byID["id-001"]; !ok || val != 0.0 {
		t.Errorf("expected ACTIVE server id-001 to have value 0.0, got %v (exists=%v)", val, ok)
	}
	if val, ok := byID["id-002"]; !ok || val != 1.0 {
		t.Errorf("expected SHUTOFF server id-002 to have value 1.0, got %v (exists=%v)", val, ok)
	}
}

func TestBuildECSNameMap(t *testing.T) {
	servers := []otcCompute.Server{
		{ID: "id-001", Name: "web-server"},
		{ID: "id-002", Name: "db-server"},
	}

	nameMap := buildECSNameMap(servers)

	if len(nameMap) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(nameMap))
	}
	if nameMap["id-001"] != "web-server" {
		t.Errorf("expected id-001 -> %q, got %q", "web-server", nameMap["id-001"])
	}
	if nameMap["id-002"] != "db-server" {
		t.Errorf("expected id-002 -> %q, got %q", "db-server", nameMap["id-002"])
	}
}

func TestMergeMetricFamilies(t *testing.T) {
	fam1 := NewGaugeMetricFamily("ecs_aom_cpu_usage", []*dto.Metric{
		NewGaugeMetric(42.0, map[string]string{"resource_id": "id-001"}),
	})
	fam2 := NewGaugeMetricFamily("ecs_aom_cpu_usage", []*dto.Metric{
		NewGaugeMetric(55.0, map[string]string{"resource_id": "id-002"}),
	})
	fam3 := NewGaugeMetricFamily("ecs_aom_memory_usage", []*dto.Metric{
		NewGaugeMetric(70.0, map[string]string{"resource_id": "id-002"}),
	})

	result := mergeMetricFamilies([]*dto.MetricFamily{fam1}, []*dto.MetricFamily{fam2, fam3})

	if len(result) != 2 {
		t.Fatalf("expected 2 families, got %d", len(result))
	}

	for _, f := range result {
		if f.GetName() == "ecs_aom_cpu_usage" {
			if len(f.Metric) != 2 {
				t.Errorf("cpu_usage: expected 2 metrics, got %d", len(f.Metric))
			}
		}
		if f.GetName() == "ecs_aom_memory_usage" {
			if len(f.Metric) != 1 {
				t.Errorf("memory_usage: expected 1 metric, got %d", len(f.Metric))
			}
		}
	}
}

func TestMergeMetricFamiliesNilInputs(t *testing.T) {
	fam := NewGaugeMetricFamily("test", []*dto.Metric{
		NewGaugeMetric(1.0, map[string]string{"id": "1"}),
	})

	// nil dst, non-nil src (first iteration of collectAOMMetrics loop).
	result := mergeMetricFamilies(nil, []*dto.MetricFamily{fam})
	if len(result) != 1 {
		t.Errorf("nil dst: expected 1 family, got %d", len(result))
	}

	// non-nil dst, nil src (server returned no AOM data).
	result = mergeMetricFamilies([]*dto.MetricFamily{fam}, nil)
	if len(result) != 1 {
		t.Errorf("nil src: expected 1 family, got %d", len(result))
	}

	// both nil.
	result = mergeMetricFamilies(nil, nil)
	if len(result) != 0 {
		t.Errorf("both nil: expected 0 families, got %d", len(result))
	}
}

func TestECSCacheIntegration(t *testing.T) {
	cache := NewNameCache()
	names := buildECSNameMap([]otcCompute.Server{
		{ID: "srv-1", Name: "web-1"},
		{ID: "srv-2", Name: "web-2"},
	})
	cache.Put("SYS.ECS", names)

	got := cache.Get("SYS.ECS")
	if len(got) != 2 {
		t.Fatalf("expected 2 cached names, got %d", len(got))
	}
	if got["srv-1"] != "web-1" {
		t.Errorf("expected srv-1 -> web-1, got %q", got["srv-1"])
	}
}
