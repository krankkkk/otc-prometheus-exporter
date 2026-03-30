package provider

import (
	"context"
	"fmt"

	cesAlarms "github.com/opentelekomcloud/gophertelekomcloud/openstack/ces/v1/alarms"
	dto "github.com/prometheus/client_model/go"

	"github.com/iits-consulting/otc-prometheus-exporter/otcclient"
)

// AlarmProvider collects CES alarm state metrics for the OTC Cloud Eye Alarm service.
// It paginates all alarm rules and reports each enabled alarm's current state as a gauge.
type AlarmProvider struct{}

func (p *AlarmProvider) Namespace() string { return "SYS.ALARM" }

func (p *AlarmProvider) Collect(ctx context.Context, client *otcclient.Client) ([]*dto.MetricFamily, error) {
	// This provider has no CES metrics — it only uses service APIs.
	// When ShouldEnrich is false (enrich=false query param), it returns no data.
	if !ShouldEnrich(ctx) {
		return nil, nil
	}

	cesClient, err := client.CESClient()
	if err != nil {
		return nil, fmt.Errorf("ces client: %w", err)
	}

	var allAlarms []cesAlarms.MetricAlarms
	var start string

	for {
		resp, err := cesAlarms.ListAlarms(cesClient, cesAlarms.ListAlarmsOpts{
			Limit: 100,
			Order: "desc",
			Start: start,
		})
		if err != nil {
			return nil, fmt.Errorf("listing CES alarms: %w", err)
		}

		allAlarms = append(allAlarms, resp.MetricAlarms...)

		// Stop paginating when we got fewer results than the limit or no marker.
		if len(resp.MetricAlarms) < 100 || resp.MetaData.Marker == "" {
			break
		}
		start = resp.MetaData.Marker
	}

	client.Logger.Debug("service API completed", "namespace", "SYS.ALARM", "alarms", len(allAlarms))

	return convertAlarmsToMetrics(allAlarms), nil
}

// alarmStateValue converts an OTC alarm state string to a numeric value:
// 0 = ok, 1 = alarm, 2 = insufficient_data
func alarmStateValue(state string) float64 {
	switch state {
	case "alarm":
		return 1.0
	case "insufficient_data":
		return 2.0
	default: // "ok"
		return 0.0
	}
}

// alarmSeverityLabel converts an OTC alarm level integer to a human-readable label:
// 1=critical, 2=major, 3=minor, 4=informational
func alarmSeverityLabel(level int) string {
	switch level {
	case 1:
		return "critical"
	case 2:
		return "major"
	case 3:
		return "minor"
	case 4:
		return "informational"
	default:
		return "unknown"
	}
}

// convertAlarmsToMetrics creates a MetricFamily for CES alarm state:
// - otc_alarm_state: 0=ok, 1=alarm, 2=insufficient_data
// Labels: alarm_name, alarm_id, namespace, metric_name, severity
func convertAlarmsToMetrics(alarms []cesAlarms.MetricAlarms) []*dto.MetricFamily {
	metrics := make([]*dto.Metric, 0, len(alarms))

	for _, a := range alarms {
		if !a.AlarmEnabled {
			continue
		}

		metrics = append(metrics, NewGaugeMetric(alarmStateValue(a.AlarmState), map[string]string{
			"alarm_name":  a.AlarmName,
			"alarm_id":    a.AlarmId,
			"namespace":   a.Metric.Namespace,
			"metric_name": a.Metric.MetricName,
			"severity":    alarmSeverityLabel(a.AlarmLevel),
		}))
	}

	if len(metrics) == 0 {
		return nil
	}

	return []*dto.MetricFamily{
		NewGaugeMetricFamily("otc_alarm_state", metrics),
	}
}

func (p *AlarmProvider) Dashboard() DashboardConfig {
	return DashboardConfig{
		Title: "ALARM - CES Alarms",
		UID:   "otc-alarm",
		Sections: []PanelSection{
			{Title: "Overview", Panels: []PanelConfig{
				{Metric: "otc_alarm_state", Title: "Firing Alarms", Unit: "short", Type: Stat,
					Expr:   `count(otc_alarm_state == 1) or vector(0)`,
					Legend: "firing",
					Thresholds: []Threshold{{Value: 0, Color: "green"}, {Value: 1, Color: "red"}}},
			}},
			{Title: "Alarm States", Panels: []PanelConfig{
				{Metric: "otc_alarm_state", Title: "All Alarms", Unit: "short", Type: Table},
			}},
		},
	}
}
