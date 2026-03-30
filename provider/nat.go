package provider

import (
	"context"

	natGW "github.com/opentelekomcloud/gophertelekomcloud/openstack/networking/v2/extensions/natgateways"
	dto "github.com/prometheus/client_model/go"

	"github.com/iits-consulting/otc-prometheus-exporter/otcclient"
)

// NATProvider collects CES metrics for the OTC NAT Gateway service,
// enriches them with gateway names, and reports gateway status.
type NATProvider struct{}

func (p *NATProvider) Namespace() string { return "SYS.NAT" }

func (p *NATProvider) Collect(ctx context.Context, client *otcclient.Client) ([]*dto.MetricFamily, error) {
	return CollectWithEnrichment(ctx, client, "SYS.NAT", func(ctx context.Context, client *otcclient.Client) (*EnrichResult, error) {
		natClient, err := client.NatV2()
		if err != nil {
			return nil, err
		}
		pages, err := natGW.List(natClient, natGW.ListOpts{}).AllPages()
		if err != nil {
			return nil, err
		}
		gateways, err := natGW.ExtractNatGateways(pages)
		if err != nil {
			return nil, err
		}
		return &EnrichResult{
			NameMap:       buildNATNameMap(gateways),
			ExtraFamilies: convertNATGatewaysToMetrics(gateways),
		}, nil
	})
}

// buildNATNameMap creates a mapping from NAT gateway ID to gateway name.
func buildNATNameMap(gateways []natGW.NatGateway) map[string]string {
	m := make(map[string]string, len(gateways))
	for _, gw := range gateways {
		m[gw.ID] = gw.Name
	}
	return m
}

// convertNATGatewaysToMetrics creates a MetricFamily "nat_gateway_status" with
// a gauge metric per gateway. The value is 0.0 for ACTIVE gateways, 1.0 otherwise
// (OTC convention: 0=normal, 1=abnormal).
func convertNATGatewaysToMetrics(gateways []natGW.NatGateway) []*dto.MetricFamily {
	metrics := make([]*dto.Metric, 0, len(gateways))
	for _, gw := range gateways {
		value := 1.0
		if gw.Status == "ACTIVE" {
			value = 0.0
		}
		metrics = append(metrics, NewGaugeMetric(value, map[string]string{
			"resource_id":   gw.ID,
			"resource_name": gw.Name,
			"status":        gw.Status,
			"spec":          gw.Spec,
		}))
	}
	return []*dto.MetricFamily{NewGaugeMetricFamily("nat_gateway_status", metrics)}
}

func (p *NATProvider) Dashboard() DashboardConfig {
	return DashboardConfig{
		Title: "NAT - NAT Gateway",
		UID:   "otc-nat",
		Sections: []PanelSection{
			{Title: "Overview", Panels: []PanelConfig{
				{Metric: "nat_gateway_status", Title: "Gateway Status", Unit: "short", Type: Stat,
					Thresholds: []Threshold{{Value: 0, Color: "green"}, {Value: 1, Color: "red"}}},
				{Metric: "nat_snat_connection", Title: "SNAT Connections", Unit: "short", Type: TimeSeries},
				{Metric: "nat_snat_connection_ratio", Title: "SNAT Connection Usage", Unit: "percent", Type: TimeSeries},
			}},
			{Title: "Bandwidth", Panels: []PanelConfig{
				{Metric: "nat_inbound_bandwidth", Title: "Inbound Bandwidth", Unit: "bps", Type: TimeSeries},
				{Metric: "nat_outbound_bandwidth", Title: "Outbound Bandwidth", Unit: "bps", Type: TimeSeries},
				{Metric: "nat_inbound_bandwidth_ratio", Title: "Inbound Bandwidth Usage", Unit: "percent", Type: TimeSeries},
				{Metric: "nat_outbound_bandwidth_ratio", Title: "Outbound Bandwidth Usage", Unit: "percent", Type: TimeSeries},
			}},
			{Title: "Packets & Traffic", Panels: []PanelConfig{
				{Metric: "nat_inbound_pps", Title: "Inbound PPS", Unit: "pps", Type: TimeSeries},
				{Metric: "nat_outbound_pps", Title: "Outbound PPS", Unit: "pps", Type: TimeSeries},
				{Metric: "nat_inbound_traffic", Title: "Inbound Traffic", Unit: "bytes", Type: TimeSeries},
				{Metric: "nat_outbound_traffic", Title: "Outbound Traffic", Unit: "bytes", Type: TimeSeries},
			}},
		},
	}
}
