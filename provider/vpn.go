package provider

import (
	"context"

	dto "github.com/prometheus/client_model/go"

	"github.com/iits-consulting/otc-prometheus-exporter/otcclient"
)

// VPNProvider collects CES metrics for the OTC Enterprise VPN service.
type VPNProvider struct{}

func (p *VPNProvider) Namespace() string { return "SYS.VPN" }

func (p *VPNProvider) Collect(ctx context.Context, client *otcclient.Client) ([]*dto.MetricFamily, error) {
	return CollectCESMetrics(ctx, client, "SYS.VPN")
}

func (p *VPNProvider) Dashboard() DashboardConfig {
	return DashboardConfig{
		Title: "VPN - Enterprise VPN",
		UID:   "otc-vpn",
		Sections: []PanelSection{
			{Title: "Connection Status", Panels: []PanelConfig{
				{Metric: "vpn_connection_status", Title: "Connection Status (0=down, 1=up, 2=unknown)", Unit: "short", Type: Stat,
					Thresholds: []Threshold{{Value: 0, Color: "red"}, {Value: 1, Color: "green"}, {Value: 2, Color: "orange"}}},
				{Metric: "vpn_bgp_peer_status", Title: "BGP Peer Status", Unit: "short", Type: Stat,
					Thresholds: []Threshold{{Value: 0, Color: "red"}, {Value: 1, Color: "green"}, {Value: 2, Color: "orange"}}},
			}},
			{Title: "Gateway Bandwidth", Panels: []PanelConfig{
				{Metric: "vpn_gateway_send_rate", Title: "Gateway Outbound Bandwidth", Unit: "bps", Type: TimeSeries},
				{Metric: "vpn_gateway_recv_rate", Title: "Gateway Inbound Bandwidth", Unit: "bps", Type: TimeSeries},
				{Metric: "vpn_gateway_send_rate_usage", Title: "Outbound Bandwidth Usage", Unit: "percent", Type: TimeSeries},
				{Metric: "vpn_gateway_recv_rate_usage", Title: "Inbound Bandwidth Usage", Unit: "percent", Type: TimeSeries},
			}},
			{Title: "Gateway Packets & Connections", Panels: []PanelConfig{
				{Metric: "vpn_gateway_send_pkt_rate", Title: "Gateway Outbound Packets", Unit: "pps", Type: TimeSeries},
				{Metric: "vpn_gateway_recv_pkt_rate", Title: "Gateway Inbound Packets", Unit: "pps", Type: TimeSeries},
				{Metric: "vpn_gateway_connection_num", Title: "Gateway Connections", Unit: "short", Type: TimeSeries},
			}},
			{Title: "Tunnel Latency & Loss", Panels: []PanelConfig{
				{Metric: "vpn_tunnel_average_latency", Title: "Tunnel Avg Latency", Unit: "ms", Type: TimeSeries},
				{Metric: "vpn_tunnel_max_latency", Title: "Tunnel Max Latency", Unit: "ms", Type: TimeSeries},
				{Metric: "vpn_tunnel_packet_loss_rate", Title: "Tunnel Packet Loss", Unit: "percent", Type: TimeSeries},
			}},
		},
	}
}
