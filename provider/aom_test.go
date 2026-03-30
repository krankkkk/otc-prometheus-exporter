package provider

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestConvertAOMResponseToFamilies(t *testing.T) {
	resp := &aomResponse{
		Metrics: []aomMetricResult{
			{
				Metric: aomMetricInfo{
					Namespace:  "PAAS.NODE",
					MetricName: "cpuUsage",
					Dimensions: []aomDimension{{Name: "hostID", Value: "host-abc"}},
				},
				Datapoints: []aomDatapoint{{
					Timestamp:  1000,
					Statistics: []aomStatistic{{Statistic: "average", Value: 42.5}},
				}},
			},
			{
				Metric: aomMetricInfo{
					Namespace:  "PAAS.NODE",
					MetricName: "diskReadRate",
					Dimensions: []aomDimension{
						{Name: "hostID", Value: "host-abc"},
						{Name: "diskDevice", Value: "vda"},
					},
				},
				Datapoints: []aomDatapoint{{
					Timestamp:  1000,
					Statistics: []aomStatistic{{Statistic: "average", Value: 1024.0}},
				}},
			},
			{
				Metric: aomMetricInfo{
					Namespace:  "PAAS.NODE",
					MetricName: "recvBytesRate",
					Dimensions: []aomDimension{
						{Name: "hostID", Value: "host-abc"},
						{Name: "netDevice", Value: "eth0"},
					},
				},
				Datapoints: []aomDatapoint{{
					Timestamp:  1000,
					Statistics: []aomStatistic{{Statistic: "average", Value: 5000.0}},
				}},
			},
			{
				Metric: aomMetricInfo{
					Namespace:  "PAAS.NODE",
					MetricName: "diskUsedRate",
					Dimensions: []aomDimension{
						{Name: "hostID", Value: "host-abc"},
						{Name: "mountPoint", Value: "/"},
						{Name: "fileSystem", Value: "ext4"},
					},
				},
				Datapoints: []aomDatapoint{{
					Timestamp:  1000,
					Statistics: []aomStatistic{{Statistic: "average", Value: 65.0}},
				}},
			},
			{
				// Empty datapoints should be skipped.
				Metric: aomMetricInfo{
					Namespace:  "PAAS.NODE",
					MetricName: "memUsedRate",
					Dimensions: []aomDimension{{Name: "hostID", Value: "host-abc"}},
				},
				Datapoints: []aomDatapoint{},
			},
			{
				// Unknown metrics are auto-named, not skipped.
				Metric: aomMetricInfo{
					Namespace:  "PAAS.NODE",
					MetricName: "someNewMetric",
					Dimensions: []aomDimension{{Name: "hostID", Value: "host-abc"}},
				},
				Datapoints: []aomDatapoint{{
					Timestamp:  1000,
					Statistics: []aomStatistic{{Statistic: "average", Value: 99.0}},
				}},
			},
		},
	}

	families := convertAOMResponseToFamilies(resp, "instance-001", "web-server")

	// 5 families: cpuUsage, diskReadRate, recvBytesRate, diskUsedRate, someNewMetric
	// (memUsedRate skipped due to empty datapoints)
	if len(families) != 5 {
		t.Fatalf("expected 5 families, got %d", len(families))
	}

	type metricData struct {
		value  float64
		labels map[string]string
	}
	famByName := make(map[string]metricData)
	for _, fam := range families {
		if len(fam.Metric) != 1 {
			t.Fatalf("family %q: expected 1 metric, got %d", fam.GetName(), len(fam.Metric))
		}
		m := fam.Metric[0]
		labels := make(map[string]string)
		for _, lp := range m.Label {
			labels[lp.GetName()] = lp.GetValue()
		}
		famByName[fam.GetName()] = metricData{value: m.Gauge.GetValue(), labels: labels}
	}

	// CPU metric.
	cpu, ok := famByName["ecs_aom_cpu_usage"]
	if !ok {
		t.Fatal("missing ecs_aom_cpu_usage family")
	}
	if cpu.value != 42.5 {
		t.Errorf("ecs_aom_cpu_usage: expected 42.5, got %f", cpu.value)
	}
	if cpu.labels["resource_id"] != "instance-001" {
		t.Errorf("expected resource_id=instance-001, got %q", cpu.labels["resource_id"])
	}
	if cpu.labels["resource_name"] != "web-server" {
		t.Errorf("expected resource_name=web-server, got %q", cpu.labels["resource_name"])
	}

	// Disk metric has disk_device label.
	disk, ok := famByName["ecs_aom_disk_read_rate"]
	if !ok {
		t.Fatal("missing ecs_aom_disk_read_rate family")
	}
	if disk.value != 1024.0 {
		t.Errorf("expected 1024.0, got %f", disk.value)
	}
	if disk.labels["disk_device"] != "vda" {
		t.Errorf("expected disk_device=vda, got %q", disk.labels["disk_device"])
	}

	// Network metric has net_device label.
	net, ok := famByName["ecs_aom_recv_bytes_rate"]
	if !ok {
		t.Fatal("missing ecs_aom_recv_bytes_rate family")
	}
	if net.labels["net_device"] != "eth0" {
		t.Errorf("expected net_device=eth0, got %q", net.labels["net_device"])
	}

	// Filesystem metric has mount_point and file_system labels.
	fs, ok := famByName["ecs_aom_disk_used_rate"]
	if !ok {
		t.Fatal("missing ecs_aom_disk_used_rate family")
	}
	if fs.labels["mount_point"] != "/" {
		t.Errorf("expected mount_point=/, got %q", fs.labels["mount_point"])
	}
	if fs.labels["file_system"] != "ext4" {
		t.Errorf("expected file_system=ext4, got %q", fs.labels["file_system"])
	}

	// Unknown metric is auto-named.
	if _, ok := famByName["ecs_aom_some_new_metric"]; !ok {
		t.Fatal("missing ecs_aom_some_new_metric — unknown metrics should be auto-named")
	}
}

func TestConvertAOMResponseFillValueFiltered(t *testing.T) {
	resp := &aomResponse{
		Metrics: []aomMetricResult{
			{
				Metric: aomMetricInfo{MetricName: "cpuUsage"},
				Datapoints: []aomDatapoint{{
					Timestamp:  1000,
					Statistics: []aomStatistic{{Statistic: "average", Value: 50.0}},
				}},
			},
			{
				// -1.0 fill value should be filtered out.
				Metric: aomMetricInfo{MetricName: "memUsedRate"},
				Datapoints: []aomDatapoint{{
					Timestamp:  1000,
					Statistics: []aomStatistic{{Statistic: "average", Value: -1.0}},
				}},
			},
			{
				// No "average" statistic should be skipped.
				Metric: aomMetricInfo{MetricName: "diskReadRate"},
				Datapoints: []aomDatapoint{{
					Timestamp:  1000,
					Statistics: []aomStatistic{{Statistic: "maximum", Value: 100.0}},
				}},
			},
		},
	}

	families := convertAOMResponseToFamilies(resp, "id", "name")

	if len(families) != 1 {
		t.Fatalf("expected 1 family (only cpuUsage), got %d", len(families))
	}
	if families[0].GetName() != "ecs_aom_cpu_usage" {
		t.Errorf("expected ecs_aom_cpu_usage, got %q", families[0].GetName())
	}
}

func TestConvertAOMResponseLatestByTimestamp(t *testing.T) {
	resp := &aomResponse{
		Metrics: []aomMetricResult{
			{
				Metric: aomMetricInfo{MetricName: "cpuUsage"},
				Datapoints: []aomDatapoint{
					{Timestamp: 3000, Statistics: []aomStatistic{{Statistic: "average", Value: 30.0}}},
					{Timestamp: 1000, Statistics: []aomStatistic{{Statistic: "average", Value: 10.0}}},
					{Timestamp: 5000, Statistics: []aomStatistic{{Statistic: "average", Value: 50.0}}},
					{Timestamp: 2000, Statistics: []aomStatistic{{Statistic: "average", Value: 20.0}}},
				},
			},
		},
	}

	families := convertAOMResponseToFamilies(resp, "id", "name")

	if len(families) != 1 {
		t.Fatalf("expected 1 family, got %d", len(families))
	}
	value := families[0].Metric[0].Gauge.GetValue()
	if value != 50.0 {
		t.Errorf("expected latest value 50.0 (timestamp 5000), got %f", value)
	}
}

func TestConvertAOMResponseNilResponse(t *testing.T) {
	families := convertAOMResponseToFamilies(nil, "id", "name")
	if families != nil {
		t.Errorf("expected nil for nil response, got %v", families)
	}
}

func TestConvertAOMResponseEmptyResponse(t *testing.T) {
	resp := &aomResponse{Metrics: []aomMetricResult{}}
	families := convertAOMResponseToFamilies(resp, "id", "name")
	if len(families) != 0 {
		t.Errorf("expected 0 families for empty response, got %d", len(families))
	}
}

func TestCheckAOMError(t *testing.T) {
	// Success code — no error.
	if err := checkAOMError("SVCSTG_AMS_2000000", "success"); err != nil {
		t.Errorf("expected nil for success code, got %v", err)
	}

	// Empty code — no error.
	if err := checkAOMError("", ""); err != nil {
		t.Errorf("expected nil for empty code, got %v", err)
	}

	// Error code — should return error.
	err := checkAOMError("SVCSTG_AMS_4000105", "Query metric data metrics is invalid")
	if err == nil {
		t.Fatal("expected error for non-success code")
	}
	if !strings.Contains(err.Error(), "SVCSTG_AMS_4000105") {
		t.Errorf("error should contain error code, got %q", err.Error())
	}
}

func TestCamelToSnake(t *testing.T) {
	tests := []struct {
		input, expected string
	}{
		{"cpuUsage", "cpu_usage"},
		{"diskReadRate", "disk_read_rate"},
		{"recvBytesRate", "recv_bytes_rate"},
		{"diskRWStatus", "disk_r_w_status"},
		{"nodeStatus", "node_status"},
		{"simple", "simple"},
		{"", ""},
		{"ABCDef", "a_b_c_def"},
	}
	for _, tt := range tests {
		got := camelToSnake(tt.input)
		if got != tt.expected {
			t.Errorf("camelToSnake(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestAomPromName(t *testing.T) {
	if got := aomPromName("cpuUsage"); got != "ecs_aom_cpu_usage" {
		t.Errorf("aomPromName(cpuUsage) = %q, want ecs_aom_cpu_usage", got)
	}
	if got := aomPromName("simple"); got != "ecs_aom_simple" {
		t.Errorf("aomPromName(simple) = %q, want ecs_aom_simple", got)
	}
}

func TestFetchAOMMetricsRespectsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately.

	_, err := fetchAOMMetrics(ctx, nil, "host-123")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
}
