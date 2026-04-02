package main

import "github.com/iits-consulting/otc-prometheus-exporter/grafana"

// allDashboardConfigs returns all dashboard configurations as static data.
func allDashboardConfigs() []grafana.DashboardConfig {
	return []grafana.DashboardConfig{
		exporterDashboard(),
		ecsDashboard(),
		rdsDashboard(),
		elbDashboard(),
		dmsDashboard(),
		natDashboard(),
		dcsDashboard(),
		ddsDashboard(),
		cbrDashboard(),
		asDashboard(),
		vpcDashboard(),
		obsDashboard(),
		wafDashboard(),
		bmsDashboard(),
		evsDashboard(),
		sfsDashboard(),
		efsDashboard(),
		dwsDashboard(),
		cssDashboard(),
		gaussdbDashboard(),
		gaussdbv5Dashboard(),
		nosqlDashboard(),
		vpnDashboard(),
		alarmDashboard(),
	}
}

func exporterDashboard() grafana.DashboardConfig {
	return grafana.DashboardConfig{
		Title: "OTC Exporter",
		UID:   "otc-exporter",
		Sections: []grafana.PanelSection{
			{Title: "Scrape Duration", Panels: []grafana.PanelConfig{
				{Metric: "otc_scrape_duration_seconds_bucket", Title: "Scrape Duration p50", Unit: "s", Type: grafana.TimeSeries,
					Expr:   `histogram_quantile(0.50, sum by (namespace, le) (rate(otc_scrape_duration_seconds_bucket[5m])))`,
					Legend: "{{namespace}}"},
				{Metric: "otc_scrape_duration_seconds_bucket", Title: "Scrape Duration p95", Unit: "s", Type: grafana.TimeSeries,
					Expr:   `histogram_quantile(0.95, sum by (namespace, le) (rate(otc_scrape_duration_seconds_bucket[5m])))`,
					Legend: "{{namespace}}"},
			}},
			{Title: "Scrape Results", Panels: []grafana.PanelConfig{
				{Metric: "otc_scrape_families_count", Title: "Families per Namespace", Unit: "short", Type: grafana.TimeSeries,
					Legend: "{{namespace}}"},
				{Metric: "otc_scrape_metrics_count", Title: "Metrics per Namespace", Unit: "short", Type: grafana.TimeSeries,
					Legend: "{{namespace}}"},
			}},
			{Title: "HTTP Request Duration", Panels: []grafana.PanelConfig{
				{Metric: "otc_http_request_duration_seconds_bucket", Title: "Request Duration p50", Unit: "s", Type: grafana.TimeSeries,
					Expr:   `histogram_quantile(0.50, sum by (host, le) (rate(otc_http_request_duration_seconds_bucket[5m])))`,
					Legend: "{{host}}"},
				{Metric: "otc_http_request_duration_seconds_bucket", Title: "Request Duration p95", Unit: "s", Type: grafana.TimeSeries,
					Expr:   `histogram_quantile(0.95, sum by (host, le) (rate(otc_http_request_duration_seconds_bucket[5m])))`,
					Legend: "{{host}}"},
			}},
			{Title: "HTTP Connection Phases", Panels: []grafana.PanelConfig{
				{Metric: "otc_http_dns_duration_seconds_bucket", Title: "DNS Lookup p95", Unit: "s", Type: grafana.TimeSeries,
					Expr:   `histogram_quantile(0.95, sum by (host, le) (rate(otc_http_dns_duration_seconds_bucket[5m])))`,
					Legend: "{{host}}"},
				{Metric: "otc_http_tls_duration_seconds_bucket", Title: "TLS Handshake p95", Unit: "s", Type: grafana.TimeSeries,
					Expr:   `histogram_quantile(0.95, sum by (host, le) (rate(otc_http_tls_duration_seconds_bucket[5m])))`,
					Legend: "{{host}}"},
				{Metric: "otc_http_ttfb_duration_seconds_bucket", Title: "Time to First Byte p95", Unit: "s", Type: grafana.TimeSeries,
					Expr:   `histogram_quantile(0.95, sum by (host, method, le) (rate(otc_http_ttfb_duration_seconds_bucket[5m])))`,
					Legend: "{{host}} {{method}}"},
			}},
			{Title: "HTTP Connections", Panels: []grafana.PanelConfig{
				{Metric: "otc_http_connections_reused_total", Title: "Connections Reused", Unit: "ops", Type: grafana.TimeSeries,
					Expr:   `rate(otc_http_connections_reused_total[5m])`,
					Legend: "{{host}}"},
				{Metric: "otc_http_connections_new_total", Title: "New Connections", Unit: "ops", Type: grafana.TimeSeries,
					Expr:   `rate(otc_http_connections_new_total[5m])`,
					Legend: "{{host}}"},
			}},
			{Title: "Go Runtime", Panels: []grafana.PanelConfig{
				{Metric: "go_goroutines", Title: "Goroutines", Unit: "short", Type: grafana.TimeSeries},
				{Metric: "go_memstats_alloc_bytes", Title: "Memory Allocated", Unit: "bytes", Type: grafana.TimeSeries},
				{Metric: "go_gc_duration_seconds", Title: "GC Duration", Unit: "s", Type: grafana.TimeSeries},
			}},
			{Title: "Process", Panels: []grafana.PanelConfig{
				{Metric: "process_cpu_seconds_total", Title: "CPU Usage", Unit: "s", Type: grafana.TimeSeries},
				{Metric: "process_resident_memory_bytes", Title: "Resident Memory", Unit: "bytes", Type: grafana.TimeSeries},
				{Metric: "process_open_fds", Title: "Open File Descriptors", Unit: "short", Type: grafana.TimeSeries},
			}},
		},
	}
}

func ecsDashboard() grafana.DashboardConfig {
	return grafana.DashboardConfig{
		Title: "ECS - Elastic Cloud Server",
		UID:   "otc-ecs",
		Sections: []grafana.PanelSection{
			{Title: "Overview", Panels: []grafana.PanelConfig{
				{Metric: "ecs_instance_status", Title: "Instance Status", Unit: "short", Type: grafana.Stat,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "green"}, {Value: 1, Color: "red"}}},
				{Metric: "ecs_aom_node_status", Title: "Host Status (AOM)", Unit: "short", Type: grafana.Stat,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "green"}, {Value: 1, Color: "red"}}},
			}},
			{Title: "CPU & Memory", Panels: []grafana.PanelConfig{
				{Metric: "ecs_cpu_util", Title: "CPU Utilization", Unit: "percent", Type: grafana.TimeSeries,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "green"}, {Value: 80, Color: "yellow"}, {Value: 90, Color: "red"}}},
				{Metric: "ecs_mem_util", Title: "Memory Utilization", Unit: "percent", Type: grafana.TimeSeries,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "green"}, {Value: 80, Color: "yellow"}, {Value: 90, Color: "red"}}},
			}},
			{Title: "Network", Panels: []grafana.PanelConfig{
				{Metric: "ecs_network_incoming_bytes_rate_inband", Title: "Inbound Traffic", Unit: "Bps", Type: grafana.TimeSeries},
				{Metric: "ecs_network_outgoing_bytes_rate_inband", Title: "Outbound Traffic", Unit: "Bps", Type: grafana.TimeSeries},
			}},
			{Title: "Disk", Panels: []grafana.PanelConfig{
				{Metric: "ecs_disk_read_bytes_rate", Title: "Disk Read", Unit: "Bps", Type: grafana.TimeSeries},
				{Metric: "ecs_disk_write_bytes_rate", Title: "Disk Write", Unit: "Bps", Type: grafana.TimeSeries},
				{Metric: "ecs_disk_read_requests_rate", Title: "Disk Read IOPS", Unit: "iops", Type: grafana.TimeSeries},
				{Metric: "ecs_disk_write_requests_rate", Title: "Disk Write IOPS", Unit: "iops", Type: grafana.TimeSeries},
			}},
			{Title: "AOM CPU & Memory", Panels: []grafana.PanelConfig{
				{Metric: "ecs_aom_cpu_usage", Title: "CPU Usage (AOM)", Unit: "percent", Type: grafana.TimeSeries,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "green"}, {Value: 80, Color: "yellow"}, {Value: 90, Color: "red"}}},
				{Metric: "ecs_aom_cpu_core_used", Title: "CPU Cores Used", Unit: "short", Type: grafana.TimeSeries},
				{Metric: "ecs_aom_mem_used_rate", Title: "Memory Usage (AOM)", Unit: "percent", Type: grafana.TimeSeries,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "green"}, {Value: 80, Color: "yellow"}, {Value: 90, Color: "red"}}},
				{Metric: "ecs_aom_free_mem", Title: "Memory Free", Unit: "decmbytes", Type: grafana.TimeSeries},
			}},
			{Title: "AOM Disk I/O (per device)", Panels: []grafana.PanelConfig{
				{Metric: "ecs_aom_disk_read_rate", Title: "Disk Read (AOM)", Unit: "KBs", Type: grafana.TimeSeries,
					Legend: "{{resource_name}} / {{disk_device}}"},
				{Metric: "ecs_aom_disk_write_rate", Title: "Disk Write (AOM)", Unit: "KBs", Type: grafana.TimeSeries,
					Legend: "{{resource_name}} / {{disk_device}}"},
			}},
			{Title: "AOM Disk Capacity (per mountpoint)", Panels: []grafana.PanelConfig{
				{Metric: "ecs_aom_disk_used_rate", Title: "Disk Usage", Unit: "percent", Type: grafana.TimeSeries,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "green"}, {Value: 80, Color: "yellow"}, {Value: 90, Color: "red"}},
					Legend:     "{{resource_name}} {{mount_point}}"},
				{Metric: "ecs_aom_disk_available_capacity", Title: "Disk Available", Unit: "decmbytes", Type: grafana.TimeSeries,
					Legend: "{{resource_name}} {{mount_point}}"},
			}},
			{Title: "AOM Network (per NIC)", Panels: []grafana.PanelConfig{
				{Metric: "ecs_aom_recv_bytes_rate", Title: "Inbound (AOM)", Unit: "Bps", Type: grafana.TimeSeries,
					Legend: "{{resource_name}} / {{net_device}}"},
				{Metric: "ecs_aom_send_bytes_rate", Title: "Outbound (AOM)", Unit: "Bps", Type: grafana.TimeSeries,
					Legend: "{{resource_name}} / {{net_device}}"},
				{Metric: "ecs_aom_recv_err_pack_rate", Title: "RX Errors", Unit: "pps", Type: grafana.TimeSeries,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "green"}, {Value: 1, Color: "red"}},
					Legend:     "{{resource_name}} / {{net_device}}"},
				{Metric: "ecs_aom_send_err_pack_rate", Title: "TX Errors", Unit: "pps", Type: grafana.TimeSeries,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "green"}, {Value: 1, Color: "red"}},
					Legend:     "{{resource_name}} / {{net_device}}"},
			}},
			{Title: "AOM Host Health", Panels: []grafana.PanelConfig{
				{Metric: "ecs_aom_process_num", Title: "Process Count", Unit: "short", Type: grafana.TimeSeries},
			}},
		},
	}
}

func rdsDashboard() grafana.DashboardConfig {
	return grafana.DashboardConfig{
		Title: "RDS - Relational Database Service",
		UID:   "otc-rds",
		Sections: []grafana.PanelSection{
			{Title: "Overview", Panels: []grafana.PanelConfig{
				{Metric: "rds_instance_status", Title: "Instance Status", Unit: "short", Type: grafana.Stat,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "green"}, {Value: 1, Color: "red"}}},
				{Metric: "rds_volume_size_gb", Title: "Volume Size", Unit: "decgbytes", Type: grafana.Gauge},
			}},
			{Title: "Node Status", Panels: []grafana.PanelConfig{
				{Metric: "rds_node_status", Title: "Node Status", Unit: "short", Type: grafana.Table,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "green"}, {Value: 1, Color: "red"}}},
			}},
			{Title: "Performance", Panels: []grafana.PanelConfig{
				{Metric: "rds_rds001_cpu_util", Title: "CPU Utilization", Unit: "percent", Type: grafana.TimeSeries,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "green"}, {Value: 80, Color: "yellow"}, {Value: 90, Color: "red"}}},
				{Metric: "rds_rds002_mem_util", Title: "Memory Utilization", Unit: "percent", Type: grafana.TimeSeries,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "green"}, {Value: 80, Color: "yellow"}, {Value: 90, Color: "red"}}},
				{Metric: "rds_rds039_disk_util", Title: "Disk Utilization", Unit: "percent", Type: grafana.TimeSeries,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "green"}, {Value: 80, Color: "yellow"}, {Value: 90, Color: "red"}}},
			}},
			{Title: "Connections", Panels: []grafana.PanelConfig{
				{Metric: "rds_rds042_database_connections", Title: "Active Connections", Unit: "short", Type: grafana.TimeSeries},
				{Metric: "rds_rds083_conn_usage", Title: "Connection Usage", Unit: "percent", Type: grafana.TimeSeries,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "green"}, {Value: 80, Color: "yellow"}, {Value: 90, Color: "red"}}},
			}},
			{Title: "Replication", Panels: []grafana.PanelConfig{
				{Metric: "rds_rds046_replication_lag", Title: "Replication Lag", Unit: "s", Type: grafana.TimeSeries,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "green"}, {Value: 10, Color: "yellow"}, {Value: 60, Color: "red"}}},
			}},
			{Title: "I/O", Panels: []grafana.PanelConfig{
				{Metric: "rds_rds003_iops", Title: "IOPS", Unit: "iops", Type: grafana.TimeSeries},
				{Metric: "rds_rds004_bytes_in", Title: "Bytes In", Unit: "Bps", Type: grafana.TimeSeries},
				{Metric: "rds_rds005_bytes_out", Title: "Bytes Out", Unit: "Bps", Type: grafana.TimeSeries},
			}},
		},
	}
}

func elbDashboard() grafana.DashboardConfig {
	return grafana.DashboardConfig{
		Title: "ELB - Elastic Load Balancer",
		UID:   "otc-elb",
		Sections: []grafana.PanelSection{
			{Title: "Overview", Panels: []grafana.PanelConfig{
				{Metric: "elb_loadbalancer_status", Title: "LB Status", Unit: "short", Type: grafana.Stat,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "green"}, {Value: 1, Color: "red"}}},
			}},
			{Title: "Traffic", Panels: []grafana.PanelConfig{
				{Metric: "elb_m1_cps", Title: "Connections/s", Unit: "cps", Type: grafana.TimeSeries},
				{Metric: "elb_m22_in_bandwidth", Title: "Inbound Bandwidth", Unit: "Bps", Type: grafana.TimeSeries},
				{Metric: "elb_m23_out_bandwidth", Title: "Outbound Bandwidth", Unit: "Bps", Type: grafana.TimeSeries},
			}},
			{Title: "Backend Health", Panels: []grafana.PanelConfig{
				{Metric: "elb_ma_normal_servers", Title: "Healthy Backends", Unit: "short", Type: grafana.TimeSeries},
				{Metric: "elb_m9_abnormal_servers", Title: "Unhealthy Backends", Unit: "short", Type: grafana.TimeSeries,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "green"}, {Value: 1, Color: "red"}}},
			}},
		},
	}
}

func dmsDashboard() grafana.DashboardConfig {
	return grafana.DashboardConfig{
		Title: "DMS - Distributed Message Service",
		UID:   "otc-dms",
		Sections: []grafana.PanelSection{
			{Title: "Overview", Panels: []grafana.PanelConfig{
				{Metric: "dms_instance_storage_used_gb", Title: "Storage Used", Unit: "decgbytes", Type: grafana.Gauge},
				{Metric: "dms_instance_storage_total_gb", Title: "Storage Total", Unit: "decgbytes", Type: grafana.Stat},
				{Metric: "dms_instance_partitions", Title: "Partitions", Unit: "short", Type: grafana.Stat},
			}},
		},
	}
}

func natDashboard() grafana.DashboardConfig {
	return grafana.DashboardConfig{
		Title: "NAT - NAT Gateway",
		UID:   "otc-nat",
		Sections: []grafana.PanelSection{
			{Title: "Overview", Panels: []grafana.PanelConfig{
				{Metric: "nat_gateway_status", Title: "Gateway Status", Unit: "short", Type: grafana.Stat,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "green"}, {Value: 1, Color: "red"}}},
				{Metric: "nat_snat_connection", Title: "SNAT Connections", Unit: "short", Type: grafana.TimeSeries},
				{Metric: "nat_snat_connection_ratio", Title: "SNAT Connection Usage", Unit: "percent", Type: grafana.TimeSeries},
			}},
			{Title: "Bandwidth", Panels: []grafana.PanelConfig{
				{Metric: "nat_inbound_bandwidth", Title: "Inbound Bandwidth", Unit: "bps", Type: grafana.TimeSeries},
				{Metric: "nat_outbound_bandwidth", Title: "Outbound Bandwidth", Unit: "bps", Type: grafana.TimeSeries},
				{Metric: "nat_inbound_bandwidth_ratio", Title: "Inbound Bandwidth Usage", Unit: "percent", Type: grafana.TimeSeries},
				{Metric: "nat_outbound_bandwidth_ratio", Title: "Outbound Bandwidth Usage", Unit: "percent", Type: grafana.TimeSeries},
			}},
			{Title: "Packets & Traffic", Panels: []grafana.PanelConfig{
				{Metric: "nat_inbound_pps", Title: "Inbound PPS", Unit: "pps", Type: grafana.TimeSeries},
				{Metric: "nat_outbound_pps", Title: "Outbound PPS", Unit: "pps", Type: grafana.TimeSeries},
				{Metric: "nat_inbound_traffic", Title: "Inbound Traffic", Unit: "bytes", Type: grafana.TimeSeries},
				{Metric: "nat_outbound_traffic", Title: "Outbound Traffic", Unit: "bytes", Type: grafana.TimeSeries},
			}},
		},
	}
}

func dcsDashboard() grafana.DashboardConfig {
	return grafana.DashboardConfig{
		Title: "DCS - Distributed Cache Service",
		UID:   "otc-dcs",
		Sections: []grafana.PanelSection{
			{Title: "Overview", Panels: []grafana.PanelConfig{
				{Metric: "dcs_instance_status", Title: "Instance Status", Unit: "short", Type: grafana.Stat,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "green"}, {Value: 1, Color: "red"}}},
				{Metric: "dcs_instance_capacity_mb", Title: "Capacity", Unit: "decmbytes", Type: grafana.Gauge},
			}},
		},
	}
}

func ddsDashboard() grafana.DashboardConfig {
	return grafana.DashboardConfig{
		Title: "DDS - Document Database Service",
		UID:   "otc-dds",
		Sections: []grafana.PanelSection{
			{Title: "Overview", Panels: []grafana.PanelConfig{
				{Metric: "dds_instance_status", Title: "Instance Status", Unit: "short", Type: grafana.Stat,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "green"}, {Value: 1, Color: "red"}}},
			}},
			{Title: "Node Status", Panels: []grafana.PanelConfig{
				{Metric: "dds_node_status", Title: "Node Status", Unit: "short", Type: grafana.Table,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "green"}, {Value: 1, Color: "red"}}},
			}},
		},
	}
}

func cbrDashboard() grafana.DashboardConfig {
	return grafana.DashboardConfig{
		Title: "CBR - Cloud Backup and Recovery",
		UID:   "otc-cbr",
		Sections: []grafana.PanelSection{
			{Title: "Backups", Panels: []grafana.PanelConfig{
				{Metric: "cbr_backup_status", Title: "Backup Status", Unit: "short", Type: grafana.Table,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "green"}, {Value: 1, Color: "red"}}},
				{Metric: "cbr_backup_size_gb", Title: "Backup Size", Unit: "decgbytes", Type: grafana.Table},
			}},
		},
	}
}

func asDashboard() grafana.DashboardConfig {
	return grafana.DashboardConfig{
		Title: "AS - Auto Scaling",
		UID:   "otc-as",
		Sections: []grafana.PanelSection{
			{Title: "Overview", Panels: []grafana.PanelConfig{
				{Metric: "as_group_status", Title: "Group Status", Unit: "short", Type: grafana.Stat,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "green"}, {Value: 1, Color: "red"}}},
			}},
			{Title: "Scaling", Panels: []grafana.PanelConfig{
				{Metric: "as_group_actual_instances", Title: "Actual Instances", Unit: "short", Type: grafana.TimeSeries},
				{Metric: "as_group_desired_instances", Title: "Desired Instances", Unit: "short", Type: grafana.TimeSeries},
				{Metric: "as_group_min_instances", Title: "Min Instances", Unit: "short", Type: grafana.TimeSeries},
				{Metric: "as_group_max_instances", Title: "Max Instances", Unit: "short", Type: grafana.TimeSeries},
			}},
		},
	}
}

func vpcDashboard() grafana.DashboardConfig {
	return grafana.DashboardConfig{
		Title: "VPC - Virtual Private Cloud",
		UID:   "otc-vpc",
		Sections: []grafana.PanelSection{
			{Title: "Overview", Panels: []grafana.PanelConfig{
				{Metric: "vpc_bandwidth_size_mbit", Title: "Bandwidth Size", Unit: "Mbits", Type: grafana.Stat},
			}},
			{Title: "Traffic", Panels: []grafana.PanelConfig{
				{Metric: "vpc_upstream_bandwidth", Title: "Upstream", Unit: "bps", Type: grafana.TimeSeries},
				{Metric: "vpc_downstream_bandwidth", Title: "Downstream", Unit: "bps", Type: grafana.TimeSeries},
			}},
		},
	}
}

func obsDashboard() grafana.DashboardConfig {
	return grafana.DashboardConfig{
		Title: "OBS - Object Storage Service",
		UID:   "otc-obs",
		Sections: []grafana.PanelSection{
			{Title: "Requests", Panels: []grafana.PanelConfig{
				{Metric: "obs_request_count_get_per_second", Title: "GET Requests/s", Unit: "reqps", Type: grafana.TimeSeries},
				{Metric: "obs_request_count_put_per_second", Title: "PUT Requests/s", Unit: "reqps", Type: grafana.TimeSeries},
				{Metric: "obs_request_count_monitor_2xx", Title: "2xx Responses", Unit: "short", Type: grafana.TimeSeries},
				{Metric: "obs_request_count_monitor_4xx", Title: "4xx Responses", Unit: "short", Type: grafana.TimeSeries,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "green"}, {Value: 1, Color: "yellow"}}},
				{Metric: "obs_request_count_monitor_5xx", Title: "5xx Responses", Unit: "short", Type: grafana.TimeSeries,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "green"}, {Value: 1, Color: "red"}}},
			}},
			{Title: "Latency", Panels: []grafana.PanelConfig{
				{Metric: "obs_request_size_le_1mb_latency_p95", Title: "Latency p95 (<1MB)", Unit: "ms", Type: grafana.TimeSeries,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "green"}, {Value: 100, Color: "yellow"}, {Value: 500, Color: "red"}}},
			}},
		},
	}
}

func wafDashboard() grafana.DashboardConfig {
	return grafana.DashboardConfig{
		Title: "WAF - Web Application Firewall",
		UID:   "otc-waf",
		Sections: []grafana.PanelSection{
			{Title: "Metrics", Panels: []grafana.PanelConfig{
				{Metric: "waf_requests", Title: "Requests", Unit: "short", Type: grafana.TimeSeries},
				{Metric: "waf_waf_http_code_2xx", Title: "2xx Responses", Unit: "short", Type: grafana.TimeSeries},
				{Metric: "waf_waf_http_code_4xx", Title: "4xx Responses", Unit: "short", Type: grafana.TimeSeries},
				{Metric: "waf_waf_http_code_5xx", Title: "5xx Responses", Unit: "short", Type: grafana.TimeSeries,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "green"}, {Value: 1, Color: "red"}}},
			}},
		},
	}
}

func bmsDashboard() grafana.DashboardConfig {
	return grafana.DashboardConfig{
		Title: "BMS - Bare Metal Server",
		UID:   "otc-bms",
		Sections: []grafana.PanelSection{
			{Title: "Metrics", Panels: []grafana.PanelConfig{
				{Metric: "bms_cpuusage", Title: "CPU Usage", Unit: "percent", Type: grafana.TimeSeries,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "green"}, {Value: 80, Color: "yellow"}, {Value: 90, Color: "red"}}},
				{Metric: "bms_memusage", Title: "Memory Usage", Unit: "percent", Type: grafana.TimeSeries,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "green"}, {Value: 80, Color: "yellow"}, {Value: 90, Color: "red"}}},
				{Metric: "bms_diskreadrate", Title: "Disk Read Rate", Unit: "Bps", Type: grafana.TimeSeries},
				{Metric: "bms_diskwriterate", Title: "Disk Write Rate", Unit: "Bps", Type: grafana.TimeSeries},
			}},
		},
	}
}

func evsDashboard() grafana.DashboardConfig {
	return grafana.DashboardConfig{
		Title: "EVS - Elastic Volume Service",
		UID:   "otc-evs",
		Sections: []grafana.PanelSection{
			{Title: "Metrics", Panels: []grafana.PanelConfig{
				{Metric: "evs_disk_read_bytes_rate", Title: "Read Throughput", Unit: "Bps", Type: grafana.TimeSeries},
				{Metric: "evs_disk_write_bytes_rate", Title: "Write Throughput", Unit: "Bps", Type: grafana.TimeSeries},
				{Metric: "evs_disk_read_requests_rate", Title: "Read IOPS", Unit: "iops", Type: grafana.TimeSeries},
				{Metric: "evs_disk_write_requests_rate", Title: "Write IOPS", Unit: "iops", Type: grafana.TimeSeries},
			}},
		},
	}
}

func sfsDashboard() grafana.DashboardConfig {
	return grafana.DashboardConfig{
		Title: "SFS - Scalable File Service",
		UID:   "otc-sfs",
		Sections: []grafana.PanelSection{
			{Title: "Metrics", Panels: []grafana.PanelConfig{
				{Metric: "sfs_read_bandwidth", Title: "Read Bandwidth", Unit: "Bps", Type: grafana.TimeSeries},
				{Metric: "sfs_write_bandwidth", Title: "Write Bandwidth", Unit: "Bps", Type: grafana.TimeSeries},
			}},
		},
	}
}

func efsDashboard() grafana.DashboardConfig {
	return grafana.DashboardConfig{
		Title: "EFS - SFS Turbo",
		UID:   "otc-efs",
		Sections: []grafana.PanelSection{
			{Title: "Metrics", Panels: []grafana.PanelConfig{
				{Metric: "efs_read_bandwidth", Title: "Read Bandwidth", Unit: "Bps", Type: grafana.TimeSeries},
				{Metric: "efs_write_bandwidth", Title: "Write Bandwidth", Unit: "Bps", Type: grafana.TimeSeries},
			}},
		},
	}
}

func dwsDashboard() grafana.DashboardConfig {
	return grafana.DashboardConfig{
		Title: "DWS - Data Warehouse Service",
		UID:   "otc-dws",
		Sections: []grafana.PanelSection{
			{Title: "Metrics", Panels: []grafana.PanelConfig{
				{Metric: "dws_cpu_usage", Title: "CPU Usage", Unit: "percent", Type: grafana.TimeSeries,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "green"}, {Value: 80, Color: "yellow"}, {Value: 90, Color: "red"}}},
				{Metric: "dws_mem_usage", Title: "Memory Usage", Unit: "percent", Type: grafana.TimeSeries,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "green"}, {Value: 80, Color: "yellow"}, {Value: 90, Color: "red"}}},
				{Metric: "dws_disk_usage", Title: "Disk Usage", Unit: "percent", Type: grafana.TimeSeries,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "green"}, {Value: 80, Color: "yellow"}, {Value: 90, Color: "red"}}},
			}},
		},
	}
}

func cssDashboard() grafana.DashboardConfig {
	return grafana.DashboardConfig{
		Title: "CSS - Cloud Search Service",
		UID:   "otc-css",
		Sections: []grafana.PanelSection{
			{Title: "Metrics", Panels: []grafana.PanelConfig{
				{Metric: "es_status", Title: "Cluster Status", Unit: "short", Type: grafana.Stat,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "green"}, {Value: 1, Color: "red"}}},
			}},
		},
	}
}

func gaussdbDashboard() grafana.DashboardConfig {
	return grafana.DashboardConfig{
		Title: "GaussDB for MySQL",
		UID:   "otc-gaussdb",
		Sections: []grafana.PanelSection{
			{Title: "Metrics", Panels: []grafana.PanelConfig{
				{Metric: "gaussdb_cpu_usage", Title: "CPU Usage", Unit: "percent", Type: grafana.TimeSeries,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "green"}, {Value: 80, Color: "yellow"}, {Value: 90, Color: "red"}}},
				{Metric: "gaussdb_mem_usage", Title: "Memory Usage", Unit: "percent", Type: grafana.TimeSeries,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "green"}, {Value: 80, Color: "yellow"}, {Value: 90, Color: "red"}}},
			}},
		},
	}
}

func gaussdbv5Dashboard() grafana.DashboardConfig {
	return grafana.DashboardConfig{
		Title: "GaussDB for openGauss",
		UID:   "otc-gaussdbv5",
		Sections: []grafana.PanelSection{
			{Title: "Metrics", Panels: []grafana.PanelConfig{
				{Metric: "gaussdbv5_cpu_usage", Title: "CPU Usage", Unit: "percent", Type: grafana.TimeSeries,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "green"}, {Value: 80, Color: "yellow"}, {Value: 90, Color: "red"}}},
				{Metric: "gaussdbv5_mem_usage", Title: "Memory Usage", Unit: "percent", Type: grafana.TimeSeries,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "green"}, {Value: 80, Color: "yellow"}, {Value: 90, Color: "red"}}},
			}},
		},
	}
}

func nosqlDashboard() grafana.DashboardConfig {
	return grafana.DashboardConfig{
		Title: "GaussDB NoSQL",
		UID:   "otc-nosql",
		Sections: []grafana.PanelSection{
			{Title: "Metrics", Panels: []grafana.PanelConfig{
				{Metric: "nosql_cpu_usage", Title: "CPU Usage", Unit: "percent", Type: grafana.TimeSeries,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "green"}, {Value: 80, Color: "yellow"}, {Value: 90, Color: "red"}}},
				{Metric: "nosql_mem_usage", Title: "Memory Usage", Unit: "percent", Type: grafana.TimeSeries,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "green"}, {Value: 80, Color: "yellow"}, {Value: 90, Color: "red"}}},
			}},
		},
	}
}

func vpnDashboard() grafana.DashboardConfig {
	return grafana.DashboardConfig{
		Title: "VPN - Enterprise VPN",
		UID:   "otc-vpn",
		Sections: []grafana.PanelSection{
			{Title: "Connection Status", Panels: []grafana.PanelConfig{
				{Metric: "vpn_connection_status", Title: "Connection Status (0=down, 1=up, 2=unknown)", Unit: "short", Type: grafana.Stat,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "red"}, {Value: 1, Color: "green"}, {Value: 2, Color: "orange"}}},
				{Metric: "vpn_bgp_peer_status", Title: "BGP Peer Status", Unit: "short", Type: grafana.Stat,
					Thresholds: []grafana.Threshold{{Value: 0, Color: "red"}, {Value: 1, Color: "green"}, {Value: 2, Color: "orange"}}},
			}},
			{Title: "Gateway Bandwidth", Panels: []grafana.PanelConfig{
				{Metric: "vpn_gateway_send_rate", Title: "Gateway Outbound Bandwidth", Unit: "bps", Type: grafana.TimeSeries},
				{Metric: "vpn_gateway_recv_rate", Title: "Gateway Inbound Bandwidth", Unit: "bps", Type: grafana.TimeSeries},
				{Metric: "vpn_gateway_send_rate_usage", Title: "Outbound Bandwidth Usage", Unit: "percent", Type: grafana.TimeSeries},
				{Metric: "vpn_gateway_recv_rate_usage", Title: "Inbound Bandwidth Usage", Unit: "percent", Type: grafana.TimeSeries},
			}},
			{Title: "Gateway Packets & Connections", Panels: []grafana.PanelConfig{
				{Metric: "vpn_gateway_send_pkt_rate", Title: "Gateway Outbound Packets", Unit: "pps", Type: grafana.TimeSeries},
				{Metric: "vpn_gateway_recv_pkt_rate", Title: "Gateway Inbound Packets", Unit: "pps", Type: grafana.TimeSeries},
				{Metric: "vpn_gateway_connection_num", Title: "Gateway Connections", Unit: "short", Type: grafana.TimeSeries},
			}},
			{Title: "Tunnel Latency & Loss", Panels: []grafana.PanelConfig{
				{Metric: "vpn_tunnel_average_latency", Title: "Tunnel Avg Latency", Unit: "ms", Type: grafana.TimeSeries},
				{Metric: "vpn_tunnel_max_latency", Title: "Tunnel Max Latency", Unit: "ms", Type: grafana.TimeSeries},
				{Metric: "vpn_tunnel_packet_loss_rate", Title: "Tunnel Packet Loss", Unit: "percent", Type: grafana.TimeSeries},
			}},
		},
	}
}

func alarmDashboard() grafana.DashboardConfig {
	return grafana.DashboardConfig{
		Title: "ALARM - CES Alarms",
		UID:   "otc-alarm",
		Sections: []grafana.PanelSection{
			{Title: "Overview", Panels: []grafana.PanelConfig{
				{Metric: "otc_alarm_state", Title: "Firing Alarms", Unit: "short", Type: grafana.Stat,
					Expr:   `count(otc_alarm_state == 1) or vector(0)`,
					Legend: "firing",
					Thresholds: []grafana.Threshold{{Value: 0, Color: "green"}, {Value: 1, Color: "red"}}},
			}},
			{Title: "Alarm States", Panels: []grafana.PanelConfig{
				{Metric: "otc_alarm_state", Title: "All Alarms", Unit: "short", Type: grafana.Table},
			}},
		},
	}
}
