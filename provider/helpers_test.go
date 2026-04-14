package provider

import (
	"context"
	"testing"

	dto "github.com/prometheus/client_model/go"
)

func TestShouldEnrichDefaultsToTrue(t *testing.T) {
	if !ShouldEnrich(context.Background()) {
		t.Error("expected ShouldEnrich to default to true")
	}
}

func TestShouldEnrichRespectsContext(t *testing.T) {
	ctx := WithEnrich(context.Background(), false)
	if ShouldEnrich(ctx) {
		t.Error("expected ShouldEnrich to return false")
	}

	ctx = WithEnrich(context.Background(), true)
	if !ShouldEnrich(ctx) {
		t.Error("expected ShouldEnrich to return true")
	}
}

func TestNewGaugeMetricFamily(t *testing.T) {
	m := NewGaugeMetric(42.5, map[string]string{"host": "server1"})
	fam := NewGaugeMetricFamily("test_metric", []*dto.Metric{m})

	if fam.GetName() != "test_metric" {
		t.Errorf("expected name %q, got %q", "test_metric", fam.GetName())
	}
	if fam.GetType() != dto.MetricType_GAUGE {
		t.Errorf("expected type GAUGE, got %v", fam.GetType())
	}
	if len(fam.Metric) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(fam.Metric))
	}
	if fam.Metric[0].Gauge.GetValue() != 42.5 {
		t.Errorf("expected gauge value 42.5, got %f", fam.Metric[0].Gauge.GetValue())
	}
}

func TestEnrichWithNames(t *testing.T) {
	m := NewGaugeMetric(1.0, map[string]string{
		"resource_id":   "abc123",
		"resource_name": "",
	})
	fam := NewGaugeMetricFamily("test_metric", []*dto.Metric{m})

	EnrichWithNames([]*dto.MetricFamily{fam}, map[string]string{
		"abc123": "my-server",
	})

	var resourceName string
	for _, lp := range fam.Metric[0].Label {
		if lp.GetName() == "resource_name" {
			resourceName = lp.GetValue()
		}
	}
	if resourceName != "my-server" {
		t.Errorf("expected resource_name %q, got %q", "my-server", resourceName)
	}
}

func TestEnrichWithNamesEmptyMap(t *testing.T) {
	m := NewGaugeMetric(1.0, map[string]string{
		"resource_id":   "abc123",
		"resource_name": "",
	})
	fam := NewGaugeMetricFamily("test_metric", []*dto.Metric{m})

	// Should not crash with nil map.
	EnrichWithNames([]*dto.MetricFamily{fam}, nil)

	// Should not crash with empty map.
	EnrichWithNames([]*dto.MetricFamily{fam}, map[string]string{})

	// resource_name should remain empty since no mapping was found.
	var resourceName string
	for _, lp := range fam.Metric[0].Label {
		if lp.GetName() == "resource_name" {
			resourceName = lp.GetValue()
		}
	}
	if resourceName != "" {
		t.Errorf("expected resource_name to remain empty, got %q", resourceName)
	}
}

func TestPrometheusMetricName(t *testing.T) {
	tests := []struct {
		namespace  string
		metricName string
		want       string
	}{
		{"SYS.ECS", "cpu_util", "ecs_cpu_util"},
		{"SERVICE.BMS", "cpuUsage", "bms_cpuusage"},
		{"SYS.DMS", "broker_data_size", "dms_broker_data_size"},
	}

	for _, tt := range tests {
		t.Run(tt.namespace+"_"+tt.metricName, func(t *testing.T) {
			got := PrometheusMetricName(tt.namespace, tt.metricName)
			if got != tt.want {
				t.Errorf("PrometheusMetricName(%q, %q) = %q, want %q",
					tt.namespace, tt.metricName, got, tt.want)
			}
		})
	}
}

func TestFillResourceNameFromLabel(t *testing.T) {
	getLabelValue := func(m *dto.Metric, name string) string {
		for _, lp := range m.Label {
			if lp.GetName() == name {
				return lp.GetValue()
			}
		}
		return ""
	}

	t.Run("copies source label when resource_name is empty", func(t *testing.T) {
		m := NewGaugeMetric(1.0, map[string]string{
			"bucket_name":   "my-bucket",
			"resource_name": "",
		})
		fam := NewGaugeMetricFamily("obs_requests", []*dto.Metric{m})
		FillResourceNameFromLabel([]*dto.MetricFamily{fam}, "bucket_name")
		if got := getLabelValue(m, "resource_name"); got != "my-bucket" {
			t.Errorf("expected resource_name %q, got %q", "my-bucket", got)
		}
	})

	t.Run("does not overwrite existing resource_name", func(t *testing.T) {
		m := NewGaugeMetric(1.0, map[string]string{
			"bucket_name":   "my-bucket",
			"resource_name": "already-set",
		})
		fam := NewGaugeMetricFamily("obs_requests", []*dto.Metric{m})
		FillResourceNameFromLabel([]*dto.MetricFamily{fam}, "bucket_name")
		if got := getLabelValue(m, "resource_name"); got != "already-set" {
			t.Errorf("expected resource_name to remain %q, got %q", "already-set", got)
		}
	})

	t.Run("no crash when source label is missing", func(t *testing.T) {
		m := NewGaugeMetric(1.0, map[string]string{
			"resource_name": "",
		})
		fam := NewGaugeMetricFamily("obs_requests", []*dto.Metric{m})
		// Should not panic; resource_name stays empty because sourceValue is "".
		FillResourceNameFromLabel([]*dto.MetricFamily{fam}, "bucket_name")
		if got := getLabelValue(m, "resource_name"); got != "" {
			t.Errorf("expected resource_name to remain empty, got %q", got)
		}
	})

	t.Run("no crash with nil families", func(t *testing.T) {
		FillResourceNameFromLabel(nil, "bucket_name")
	})

	t.Run("no crash with empty families", func(t *testing.T) {
		FillResourceNameFromLabel([]*dto.MetricFamily{}, "bucket_name")
	})
}

func TestNewGaugeMetricWithTimestamp(t *testing.T) {
	var ts int64 = 1700000000000
	m := NewGaugeMetricWithTimestamp(3.14, map[string]string{
		"region": "eu-de",
		"host":   "server1",
	}, ts)

	if m.TimestampMs == nil {
		t.Fatal("expected TimestampMs to be set, got nil")
	}
	if *m.TimestampMs != ts {
		t.Errorf("expected TimestampMs %d, got %d", ts, *m.TimestampMs)
	}

	if m.Gauge == nil {
		t.Fatal("expected Gauge to be set, got nil")
	}
	if m.Gauge.GetValue() != 3.14 {
		t.Errorf("expected gauge value 3.14, got %f", m.Gauge.GetValue())
	}

	getLabelValue := func(name string) string {
		for _, lp := range m.Label {
			if lp.GetName() == name {
				return lp.GetValue()
			}
		}
		return ""
	}
	if got := getLabelValue("region"); got != "eu-de" {
		t.Errorf("expected label region=%q, got %q", "eu-de", got)
	}
	if got := getLabelValue("host"); got != "server1" {
		t.Errorf("expected label host=%q, got %q", "server1", got)
	}
}
