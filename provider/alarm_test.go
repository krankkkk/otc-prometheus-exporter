package provider

import (
	"testing"

	cesAlarms "github.com/opentelekomcloud/gophertelekomcloud/openstack/ces/v1/alarms"
)

func TestConvertAlarmsToMetrics(t *testing.T) {
	alarms := []cesAlarms.MetricAlarms{
		{
			AlarmName:    "high-cpu",
			AlarmId:      "al-001",
			AlarmEnabled: true,
			AlarmLevel:   1,
			AlarmState:   "alarm",
			Metric: cesAlarms.MetricForAlarm{
				Namespace:  "SYS.ECS",
				MetricName: "cpu_util",
			},
		},
		{
			AlarmName:    "low-memory",
			AlarmId:      "al-002",
			AlarmEnabled: true,
			AlarmLevel:   2,
			AlarmState:   "ok",
			Metric: cesAlarms.MetricForAlarm{
				Namespace:  "SYS.ECS",
				MetricName: "mem_util",
			},
		},
		{
			AlarmName:    "disk-io",
			AlarmId:      "al-003",
			AlarmEnabled: true,
			AlarmLevel:   3,
			AlarmState:   "insufficient_data",
			Metric: cesAlarms.MetricForAlarm{
				Namespace:  "SYS.ECS",
				MetricName: "disk_util",
			},
		},
		{
			AlarmName:    "disabled-alarm",
			AlarmId:      "al-004",
			AlarmEnabled: false,
			AlarmLevel:   4,
			AlarmState:   "alarm",
			Metric: cesAlarms.MetricForAlarm{
				Namespace:  "SYS.RDS",
				MetricName: "rds_cpu_util",
			},
		},
	}

	families := convertAlarmsToMetrics(alarms)

	// Should have exactly 1 family.
	if len(families) != 1 {
		t.Fatalf("expected 1 family, got %d", len(families))
	}

	fam := families[0]

	// Family name should be otc_alarm_state.
	if fam.GetName() != "otc_alarm_state" {
		t.Errorf("expected family name %q, got %q", "otc_alarm_state", fam.GetName())
	}

	// Should have 3 metrics (disabled alarm skipped).
	if len(fam.Metric) != 3 {
		t.Fatalf("expected 3 metrics (disabled skipped), got %d", len(fam.Metric))
	}

	// Build a map from alarm_id to metric for easy assertion.
	byID := make(map[string]*struct {
		value  float64
		labels map[string]string
	})
	for _, m := range fam.Metric {
		labels := make(map[string]string)
		for _, lp := range m.Label {
			labels[lp.GetName()] = lp.GetValue()
		}
		id := labels["alarm_id"]
		val := m.Gauge.GetValue()
		byID[id] = &struct {
			value  float64
			labels map[string]string
		}{value: val, labels: labels}
	}

	// al-004 (disabled) must not appear.
	if _, ok := byID["al-004"]; ok {
		t.Error("disabled alarm al-004 should not appear in metrics")
	}

	// al-001: state=alarm -> value 1.0, severity=critical.
	m1, ok := byID["al-001"]
	if !ok {
		t.Fatal("alarm al-001 not found in metrics")
	}
	if m1.value != 1.0 {
		t.Errorf("al-001: expected value 1.0 (alarm), got %f", m1.value)
	}
	if m1.labels["severity"] != "critical" {
		t.Errorf("al-001: expected severity %q, got %q", "critical", m1.labels["severity"])
	}
	if m1.labels["alarm_name"] != "high-cpu" {
		t.Errorf("al-001: expected alarm_name %q, got %q", "high-cpu", m1.labels["alarm_name"])
	}
	if m1.labels["namespace"] != "SYS.ECS" {
		t.Errorf("al-001: expected namespace %q, got %q", "SYS.ECS", m1.labels["namespace"])
	}
	if m1.labels["metric_name"] != "cpu_util" {
		t.Errorf("al-001: expected metric_name %q, got %q", "cpu_util", m1.labels["metric_name"])
	}

	// al-002: state=ok -> value 0.0, severity=major.
	m2, ok := byID["al-002"]
	if !ok {
		t.Fatal("alarm al-002 not found in metrics")
	}
	if m2.value != 0.0 {
		t.Errorf("al-002: expected value 0.0 (ok), got %f", m2.value)
	}
	if m2.labels["severity"] != "major" {
		t.Errorf("al-002: expected severity %q, got %q", "major", m2.labels["severity"])
	}

	// al-003: state=insufficient_data -> value 2.0, severity=minor.
	m3, ok := byID["al-003"]
	if !ok {
		t.Fatal("alarm al-003 not found in metrics")
	}
	if m3.value != 2.0 {
		t.Errorf("al-003: expected value 2.0 (insufficient_data), got %f", m3.value)
	}
	if m3.labels["severity"] != "minor" {
		t.Errorf("al-003: expected severity %q, got %q", "minor", m3.labels["severity"])
	}
}

func TestConvertAlarmsToMetrics_Empty(t *testing.T) {
	families := convertAlarmsToMetrics([]cesAlarms.MetricAlarms{})
	if families != nil {
		t.Errorf("expected nil for empty input, got %v", families)
	}
}

func TestConvertAlarmsToMetrics_AllDisabled(t *testing.T) {
	alarms := []cesAlarms.MetricAlarms{
		{
			AlarmName:    "disabled-1",
			AlarmId:      "al-d01",
			AlarmEnabled: false,
			AlarmState:   "alarm",
			Metric:       cesAlarms.MetricForAlarm{Namespace: "SYS.ECS", MetricName: "cpu_util"},
		},
		{
			AlarmName:    "disabled-2",
			AlarmId:      "al-d02",
			AlarmEnabled: false,
			AlarmState:   "ok",
			Metric:       cesAlarms.MetricForAlarm{Namespace: "SYS.RDS", MetricName: "rds_cpu_util"},
		},
	}

	families := convertAlarmsToMetrics(alarms)
	if families != nil {
		t.Errorf("expected nil when all alarms are disabled, got %v", families)
	}
}

func TestAlarmSeverityLabel(t *testing.T) {
	tests := []struct {
		level    int
		expected string
	}{
		{1, "critical"},
		{2, "major"},
		{3, "minor"},
		{4, "informational"},
		{0, "unknown"},
		{99, "unknown"},
	}
	for _, tc := range tests {
		got := alarmSeverityLabel(tc.level)
		if got != tc.expected {
			t.Errorf("alarmSeverityLabel(%d): expected %q, got %q", tc.level, tc.expected, got)
		}
	}
}

func TestAlarmStateValue(t *testing.T) {
	tests := []struct {
		state    string
		expected float64
	}{
		{"ok", 0.0},
		{"alarm", 1.0},
		{"insufficient_data", 2.0},
		{"unknown", 0.0}, // defaults to ok
	}
	for _, tc := range tests {
		got := alarmStateValue(tc.state)
		if got != tc.expected {
			t.Errorf("alarmStateValue(%q): expected %f, got %f", tc.state, tc.expected, got)
		}
	}
}
