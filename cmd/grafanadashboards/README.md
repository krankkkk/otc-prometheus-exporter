# Grafana Dashboard Generator

Generates Grafana dashboards from provider-defined metadata.

## Running

```shell
go run ./cmd/grafanadashboards/ --output-path ./charts/otc-prometheus-exporter/dashboards
```

Or via Make:

```shell
make generate-dashboards-into-helm
```
