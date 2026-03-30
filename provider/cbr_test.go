package provider

import (
	"testing"

	cbrBackups "github.com/opentelekomcloud/gophertelekomcloud/openstack/cbr/v3/backups"
)

func TestConvertCBRBackupsToMetrics(t *testing.T) {
	backups := []cbrBackups.Backup{
		{
			ID:           "bk-001",
			Name:         "daily-backup",
			Status:       "available",
			ResourceSize: 100,
			ResourceID:   "res-001",
			ResourceName: "my-server",
		},
		{
			ID:           "bk-002",
			Name:         "weekly-backup",
			Status:       "error",
			ResourceSize: 50,
			ResourceID:   "res-002",
			ResourceName: "my-database",
		},
	}

	families := convertCBRBackupsToMetrics(backups)

	if len(families) != 2 {
		t.Fatalf("expected 2 families, got %d", len(families))
	}

	// Verify family names.
	expectedNames := []string{"cbr_backup_status", "cbr_backup_size_gb"}
	for i, name := range expectedNames {
		if families[i].GetName() != name {
			t.Errorf("expected family[%d] name %q, got %q", i, name, families[i].GetName())
		}
	}

	// Each family should have 2 metrics (one per backup).
	for i, fam := range families {
		if len(fam.Metric) != 2 {
			t.Errorf("family[%d] %q: expected 2 metrics, got %d", i, fam.GetName(), len(fam.Metric))
		}
	}

	// Verify status values: available=0.0, error=1.0 (OTC convention: 0=normal, 1=abnormal).
	statusFam := families[0]
	byID := make(map[string]float64)
	for _, m := range statusFam.Metric {
		for _, lp := range m.Label {
			if lp.GetName() == "backup_id" {
				byID[lp.GetValue()] = m.Gauge.GetValue()
			}
		}
	}
	if val, ok := byID["bk-001"]; !ok || val != 0.0 {
		t.Errorf("expected available backup bk-001 to have value 0.0, got %v (exists=%v)", val, ok)
	}
	if val, ok := byID["bk-002"]; !ok || val != 1.0 {
		t.Errorf("expected error backup bk-002 to have value 1.0, got %v (exists=%v)", val, ok)
	}

	// Verify size values.
	sizeFam := families[1]
	sizeByID := make(map[string]float64)
	for _, m := range sizeFam.Metric {
		for _, lp := range m.Label {
			if lp.GetName() == "backup_id" {
				sizeByID[lp.GetValue()] = m.Gauge.GetValue()
			}
		}
	}
	if val, ok := sizeByID["bk-001"]; !ok || val != 100.0 {
		t.Errorf("expected backup bk-001 size 100.0, got %v (exists=%v)", val, ok)
	}
	if val, ok := sizeByID["bk-002"]; !ok || val != 50.0 {
		t.Errorf("expected backup bk-002 size 50.0, got %v (exists=%v)", val, ok)
	}
}
