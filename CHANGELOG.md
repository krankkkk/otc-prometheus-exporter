# Changelog

## v1.0.0 — Complete Rewrite

This is a full rewrite of the exporter. The architecture changed fundamentally — please read the breaking changes carefully before upgrading.

### Breaking Changes

#### Scraping model changed: poll&cache → on-demand-pull

The old exporter ran a background loop that fetched all namespaces on a fixed interval (`WAITDURATION`) and served cached metrics. The new exporter fetches metrics **on demand** — each Prometheus scrape triggers a real-time API call for exactly the requested namespace via `?namespace=SYS.XXX`.

#### Removed environment variables

| Removed | Replacement |
|---|---|
| `NAMESPACES` | No longer needed — Prometheus scrape config controls which namespaces are fetched |
| `WAITDURATION` | No longer needed — Prometheus scrape interval controls fetch frequency |
| `FETCH_RESOURCE_ID_TO_NAME` | Replaced by `?enrich=false` query parameter (default: enrichment enabled) |

#### New / renamed environment variables

| Variable | Default | Description |
|---|---|---|
| `REGION` | `eu-de` | **New.** OTC region (`eu-de` or `eu-nl`) |
| `OS_REGION_PROJECT_ID` | auto-discovered | **New.** Region-level project ID for global services (OBS). Auto-discovered with user/pass auth. |
| `REQUEST_TIMEOUT` | `10` | **New.** HTTP timeout in seconds for OTC API calls |
| `IDLE_CONN_TIMEOUT` | `90` | **New.** Idle HTTP connection pool timeout in seconds |
| `COLLECT_TIMEOUT` | `55` | **New.** Max collect time per scrape in seconds |
| `CES_BATCH_SIZE` | `500` | **New.** Max metrics per CES batch API request |
| `CES_LOOKBACK` | `10` | **New.** CES lookback window in minutes |
| `AOM_BATCH_SIZE` | `20` | **New.** Max metrics per AOM data API request |
| `AOM_CONCURRENCY` | `5` | **New.** Max concurrent AOM API calls per scrape |
| `LOG_LEVEL` | `INFO` | **New.** Log level (DEBUG, INFO, WARN, ERROR) |

#### Helm chart: ServiceMonitor → PodMonitor

The chart no longer creates a `Service` + `ServiceMonitor`. It now uses a **PodMonitor** with one `podMetricsEndpoint` per enabled namespace. If you have custom `ServiceMonitor` configurations, migrate them to the new `podMonitor` section in values.yaml.

#### Helm chart: values.yaml restructured

The values.yaml structure changed significantly:

| Old | New |
|---|---|
| `service` | Removed — no Service resource created |
| `serviceMonitor` | `podMonitor` (with `selfMonitoring` toggle) |
| `serviceMonitor.namespaces` | `targets.namespaces` (unified config for scraping, dashboards, and alerts) |
| `prometheusRules.rules.rds` | `targets.namespaces.SYS.RDS.rules: true` (per-namespace) |
| `prometheusRules.rules.elb` | `targets.namespaces.SYS.ELB.rules: true` (per-namespace) |
| `prometheusRules.rules.obs` | `targets.namespaces.SYS.OBS.rules: true` (per-namespace) |
| `prometheusRules.rules.alarm` | `targets.namespaces.SYS.ALARM.rules: true` (per-namespace) |

#### Metric names changed

Metric names now follow a consistent `{namespace}_{metric}` pattern derived from the CES namespace and metric name. Old custom metric names are no longer produced. Check your dashboards and alert rules for references to old metric names.

### New Features

#### Per-namespace scraping

Each OTC namespace is scraped independently via `?namespace=SYS.XXX`. This means:
- Individual scrape intervals and timeouts per namespace
- One failing namespace doesn't block others
- Prometheus shows per-namespace scrape health (up/down)

#### Service-API enrichment

Providers that call service-specific APIs (ECS, RDS, ELB, DMS, NAT, DCS, DDS, VPC, CBR, AS) now enrich CES metrics with human-readable resource names and additional status/capacity metrics. Disable per-scrape with `?enrich=false`.

#### AOM host metrics (ECS)

When enrichment is enabled, ECS scrapes also fetch per-device AOM metrics (CPU, memory, disk I/O per device, network per NIC, filesystem per mountpoint) under the `ecs_aom_` prefix. Requires `AOM ReadOnlyAccess` and ICAgent installed in guest OS.

#### New services

Added providers for: VPN (`SYS.VPN`), CES Alarms (`SYS.ALARM`), Auto Scaling status metrics (`SYS.AS`), CBR backup status (`SYS.CBR`).

#### Grafana dashboards

Pre-generated dashboards for all 24 namespaces, deployed as ConfigMaps with `grafana_dashboard: "1"` label. Includes a self-monitoring dashboard for the exporter itself (HTTP trace metrics, scrape durations).

#### PrometheusRule alerts

Bundled alert rules for RDS, ELB, OBS, and Alarms, deployed as PrometheusRule CRDs. Enable per namespace via `targets.namespaces.SYS.XXX.rules: true`.
