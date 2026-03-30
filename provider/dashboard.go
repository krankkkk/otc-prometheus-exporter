package provider

// DashboardProvider is implemented by providers that want a Grafana dashboard generated.
type DashboardProvider interface {
	Dashboard() DashboardConfig
}

type DashboardConfig struct {
	Title    string
	UID      string
	Sections []PanelSection
}

type PanelSection struct {
	Title  string
	Panels []PanelConfig
}

type PanelConfig struct {
	Metric     string
	Title      string
	Unit       string
	Type       PanelType
	Thresholds []Threshold
	Legend     string // Override legend format (default: "{{resource_name}}")
	Expr       string // Override PromQL expression (default: Metric{resource_name=~"$resource_name"})
}

type PanelType int

const (
	TimeSeries PanelType = iota
	Stat
	Gauge
	Table
)

type Threshold struct {
	Value float64
	Color string
}
