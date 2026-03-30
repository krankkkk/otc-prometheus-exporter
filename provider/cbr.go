package provider

import (
	"context"
	"fmt"

	cbrBackups "github.com/opentelekomcloud/gophertelekomcloud/openstack/cbr/v3/backups"
	dto "github.com/prometheus/client_model/go"

	"github.com/iits-consulting/otc-prometheus-exporter/otcclient"
)

// CBRProvider collects service-API metrics for the OTC Cloud Backup and Recovery service.
// CES has no CBR metrics, so this provider only reports backup status and size.
type CBRProvider struct{}

func (p *CBRProvider) Namespace() string { return "SYS.CBR" }

func (p *CBRProvider) Collect(ctx context.Context, client *otcclient.Client) ([]*dto.MetricFamily, error) {
	// This provider has no CES metrics — it only uses service APIs.
	// When ShouldEnrich is false (enrich=false query param), it returns no data.
	if !ShouldEnrich(ctx) {
		return nil, nil
	}

	cbrClient, err := client.CBRV3()
	if err != nil {
		return nil, fmt.Errorf("cbr client: %w", err)
	}

	backups, err := cbrBackups.List(cbrClient, cbrBackups.ListOpts{})
	if err != nil {
		return nil, fmt.Errorf("listing CBR backups: %w", err)
	}

	client.Logger.Debug("service API completed", "namespace", "SYS.CBR", "backups", len(backups))

	return convertCBRBackupsToMetrics(backups), nil
}

// convertCBRBackupsToMetrics creates MetricFamily objects for CBR-specific metrics:
// - cbr_backup_status: 0.0 if available, 1.0 otherwise (OTC convention: 0=normal, 1=abnormal)
// - cbr_backup_size_gb: backup resource size in GB
func convertCBRBackupsToMetrics(backups []cbrBackups.Backup) []*dto.MetricFamily {
	statusMetrics := make([]*dto.Metric, 0, len(backups))
	sizeMetrics := make([]*dto.Metric, 0, len(backups))

	for _, bk := range backups {
		statusValue := 1.0
		if bk.Status == "available" {
			statusValue = 0.0
		}
		statusMetrics = append(statusMetrics, NewGaugeMetric(statusValue, map[string]string{
			"backup_id":     bk.ID,
			"backup_name":   bk.Name,
			"resource_id":   bk.ResourceID,
			"resource_name": bk.ResourceName,
			"status":        bk.Status,
		}))

		sizeMetrics = append(sizeMetrics, NewGaugeMetric(float64(bk.ResourceSize), map[string]string{
			"backup_id":     bk.ID,
			"resource_id":   bk.ResourceID,
			"resource_name": bk.ResourceName,
		}))
	}

	return []*dto.MetricFamily{
		NewGaugeMetricFamily("cbr_backup_status", statusMetrics),
		NewGaugeMetricFamily("cbr_backup_size_gb", sizeMetrics),
	}
}

func (p *CBRProvider) Dashboard() DashboardConfig {
	return DashboardConfig{
		Title: "CBR - Cloud Backup and Recovery",
		UID:   "otc-cbr",
		Sections: []PanelSection{
			{Title: "Backups", Panels: []PanelConfig{
				{Metric: "cbr_backup_status", Title: "Backup Status", Unit: "short", Type: Table,
					Thresholds: []Threshold{{Value: 0, Color: "green"}, {Value: 1, Color: "red"}}},
				{Metric: "cbr_backup_size_gb", Title: "Backup Size", Unit: "decgbytes", Type: Table},
			}},
		},
	}
}
