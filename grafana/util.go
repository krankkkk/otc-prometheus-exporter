package grafana

// NewDefaultDashboard creates a base Grafana dashboard with standard OTC settings.
func NewDefaultDashboard(title, uid string) Dashboard {
	return Dashboard{
		Inputs: []Input{
			{
				Name:        "DS_PROMETHEUS",
				Label:       "Prometheus",
				Description: "",
				Type:        "datasource",
				PluginID:    "prometheus",
				PluginName:  "Prometheus",
			},
		},
		Elements: Elements{},
		Requires: []Require{
			{Type: "grafana", ID: "grafana", Name: "Grafana", Version: "10.3.3"},
			{Type: "datasource", ID: "prometheus", Name: "Prometheus", Version: "1.0.0"},
			{Type: "panel", ID: "timeseries", Name: "Time series"},
			{Type: "panel", ID: "stat", Name: "Stat"},
			{Type: "panel", ID: "gauge", Name: "Gauge"},
			{Type: "panel", ID: "table", Name: "Table"},
		},
		Annotations: Annotations{
			List: []AnnotationList{{
				BuiltIn:    1,
				Datasource: Datasource{Type: "grafana", UID: "-- Grafana --"},
				Enable:     true,
				Hide:       true,
				IconColor:  "rgba(0, 211, 255, 1)",
				Name:       "Annotations & Alerts",
				Type:       "dashboard",
			}},
		},
		Editable:             true,
		FiscalYearStartMonth: 0,
		GraphTooltip:         0,
		ID:                   nil,
		Links:                []any{},
		LiveNow:              false,
		Panels:               []Panel{},
		Refresh:              "30s",
		SchemaVersion:        39,
		Tags:                 []string{"OTC"},
		Templating: Templating{
			List: []TemplatingVariable{
				{
					Current:     Current{Selected: true, Text: "Prometheus", Value: "prometheus"},
					Description: "Datasource where the OTC Prometheus Exporter data is stored",
					Hide:        0,
					IncludeAll:  false,
					Label:       "Datasource",
					Multi:       false,
					Name:        "datasource",
					Options:     []any{},
					Query:       "prometheus",
					Refresh:     1,
					SkipURLSync: false,
					Type:        "datasource",
				},
			},
		},
		Time:       Time{From: "now-24h", To: "now"},
		Timepicker: Timepicker{},
		Timezone:   "browser",
		Title:      title,
		UID:        uid,
		Version:    4,
		WeekStart:  "",
	}
}
