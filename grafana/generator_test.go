package grafana

import (
	"testing"
)

func TestGenerateDashboard(t *testing.T) {
	cfg := DashboardConfig{
		Title: "Test Dashboard",
		UID:   "test-uid",
		Sections: []PanelSection{
			{
				Title: "Overview",
				Panels: []PanelConfig{
					{Metric: "test_status", Title: "Status", Unit: "short", Type: Stat,
						Thresholds: []Threshold{{Value: 0, Color: "red"}, {Value: 1, Color: "green"}}},
				},
			},
			{
				Title: "Performance",
				Panels: []PanelConfig{
					{Metric: "test_cpu", Title: "CPU", Unit: "percent", Type: TimeSeries,
						Thresholds: []Threshold{{Value: 0, Color: "green"}, {Value: 80, Color: "yellow"}, {Value: 90, Color: "red"}}},
					{Metric: "test_mem", Title: "Memory", Unit: "percent", Type: TimeSeries,
						Thresholds: []Threshold{{Value: 0, Color: "green"}, {Value: 80, Color: "red"}}},
				},
			},
		},
	}

	board := GenerateDashboard(cfg)

	if board.Title != "Test Dashboard" {
		t.Errorf("expected title 'Test Dashboard', got %q", board.Title)
	}
	if board.UID != "test-uid" {
		t.Errorf("expected UID 'test-uid', got %q", board.UID)
	}

	// 2 row panels + 3 metric panels = 5 total
	if len(board.Panels) != 5 {
		t.Fatalf("expected 5 panels (2 rows + 3 metrics), got %d", len(board.Panels))
	}

	if board.Panels[0].Type != "row" {
		t.Errorf("expected first panel to be row, got %q", board.Panels[0].Type)
	}
	if board.Panels[0].Title != "Overview" {
		t.Errorf("expected row title 'Overview', got %q", board.Panels[0].Title)
	}
	if board.Panels[1].Type != "stat" {
		t.Errorf("expected stat panel, got %q", board.Panels[1].Type)
	}
	if board.Panels[2].Type != "row" {
		t.Errorf("expected row panel, got %q", board.Panels[2].Type)
	}
	if board.Panels[3].Type != "timeseries" {
		t.Errorf("expected timeseries panel, got %q", board.Panels[3].Type)
	}
}

func TestPanelTypes(t *testing.T) {
	tests := []struct {
		panelType PanelType
		expected  string
	}{
		{TimeSeries, "timeseries"},
		{Stat, "stat"},
		{Gauge, "gauge"},
		{Table, "table"},
	}

	for _, tt := range tests {
		cfg := DashboardConfig{
			Title: "Test", UID: "test",
			Sections: []PanelSection{{
				Title: "Section",
				Panels: []PanelConfig{{
					Metric: "m", Title: "T", Unit: "short", Type: tt.panelType,
				}},
			}},
		}
		board := GenerateDashboard(cfg)
		if len(board.Panels) < 2 {
			t.Fatalf("expected at least 2 panels, got %d", len(board.Panels))
		}
		if board.Panels[1].Type != tt.expected {
			t.Errorf("PanelType %d: expected %q, got %q", tt.panelType, tt.expected, board.Panels[1].Type)
		}
	}
}

func TestTablePanelsNonOverlapping(t *testing.T) {
	// Two Table panels in the same section must get distinct y values.
	cfg := DashboardConfig{
		Title: "Test", UID: "test",
		Sections: []PanelSection{{
			Title: "Status",
			Panels: []PanelConfig{
				{Metric: "svc_backup_status", Title: "Backup Status", Unit: "short", Type: Table},
				{Metric: "svc_backup_size", Title: "Backup Size", Unit: "decgbytes", Type: Table},
			},
		}},
	}
	board := GenerateDashboard(cfg)

	// 1 row panel + 2 table panels = 3 total
	if len(board.Panels) != 3 {
		t.Fatalf("expected 3 panels, got %d", len(board.Panels))
	}

	p1 := board.Panels[1]
	p2 := board.Panels[2]

	if p1.Type != "table" {
		t.Errorf("expected panel[1] to be table, got %q", p1.Type)
	}
	if p2.Type != "table" {
		t.Errorf("expected panel[2] to be table, got %q", p2.Type)
	}
	if p1.GridPos.Y == p2.GridPos.Y {
		t.Errorf("two Table panels have the same Y=%d; they overlap", p1.GridPos.Y)
	}
	if p2.GridPos.Y != p1.GridPos.Y+p1.GridPos.H {
		t.Errorf("expected panel[2].Y=%d (panel[1].Y + panel[1].H), got %d",
			p1.GridPos.Y+p1.GridPos.H, p2.GridPos.Y)
	}
	// Both must be full-width
	if p1.GridPos.W != 24 {
		t.Errorf("expected panel[1].W=24, got %d", p1.GridPos.W)
	}
	if p2.GridPos.W != 24 {
		t.Errorf("expected panel[2].W=24, got %d", p2.GridPos.W)
	}
}

func TestMixedFullAndHalfWidthPanels(t *testing.T) {
	// A Table panel followed by two half-width panels must not overlap.
	cfg := DashboardConfig{
		Title: "Test", UID: "test",
		Sections: []PanelSection{{
			Title: "Mixed",
			Panels: []PanelConfig{
				{Metric: "svc_status", Title: "Status", Unit: "short", Type: Table},
				{Metric: "svc_cpu", Title: "CPU", Unit: "percent", Type: TimeSeries},
				{Metric: "svc_mem", Title: "Memory", Unit: "percent", Type: TimeSeries},
			},
		}},
	}
	board := GenerateDashboard(cfg)

	// 1 row + 3 metric panels = 4 total
	if len(board.Panels) != 4 {
		t.Fatalf("expected 4 panels, got %d", len(board.Panels))
	}

	tablePanel := board.Panels[1]
	half1 := board.Panels[2]
	half2 := board.Panels[3]

	if tablePanel.Type != "table" {
		t.Errorf("expected panels[1] to be table, got %q", tablePanel.Type)
	}

	// The two half-width panels must start at y = tablePanel.Y + tablePanel.H
	expectedHalfY := tablePanel.GridPos.Y + tablePanel.GridPos.H
	if half1.GridPos.Y != expectedHalfY {
		t.Errorf("half panel 1: expected Y=%d, got %d", expectedHalfY, half1.GridPos.Y)
	}
	if half2.GridPos.Y != expectedHalfY {
		t.Errorf("half panel 2: expected Y=%d, got %d", expectedHalfY, half2.GridPos.Y)
	}

	// The two half-width panels must be side by side (x=0 and x=12)
	if half1.GridPos.X != 0 {
		t.Errorf("half panel 1: expected X=0, got %d", half1.GridPos.X)
	}
	if half2.GridPos.X != 12 {
		t.Errorf("half panel 2: expected X=12, got %d", half2.GridPos.X)
	}
}

func TestThresholdsConverted(t *testing.T) {
	cfg := DashboardConfig{
		Title: "Test", UID: "test",
		Sections: []PanelSection{{
			Title: "S",
			Panels: []PanelConfig{{
				Metric: "m", Title: "T", Unit: "percent", Type: TimeSeries,
				Thresholds: []Threshold{
					{Value: 0, Color: "green"},
					{Value: 80, Color: "yellow"},
					{Value: 90, Color: "red"},
				},
			}},
		}},
	}
	board := GenerateDashboard(cfg)
	steps := board.Panels[1].FieldConfig.Defaults.Thresholds.Steps
	if len(steps) != 3 {
		t.Fatalf("expected 3 threshold steps, got %d", len(steps))
	}
	if steps[0].Color != "green" {
		t.Errorf("expected first step green, got %s", steps[0].Color)
	}
	if steps[2].Color != "red" {
		t.Errorf("expected last step red, got %s", steps[2].Color)
	}
}
