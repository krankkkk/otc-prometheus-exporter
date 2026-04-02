package provider

import (
	"context"
	"sync"

	otcCompute "github.com/opentelekomcloud/gophertelekomcloud/openstack/compute/v2/servers"
	dto "github.com/prometheus/client_model/go"

	"github.com/iits-consulting/otc-prometheus-exporter/otcclient"
)

// ECSProvider collects CES metrics for the OTC Elastic Cloud Server service,
// enriches them with server names, and reports instance status.
type ECSProvider struct{}

func (p *ECSProvider) Namespace() string { return "SYS.ECS" }

func (p *ECSProvider) Collect(ctx context.Context, client *otcclient.Client) ([]*dto.MetricFamily, error) {
	return CollectWithEnrichment(ctx, client, "SYS.ECS", func(ctx context.Context, client *otcclient.Client) (*EnrichResult, error) {
		computeClient, err := client.ComputeV2()
		if err != nil {
			return nil, err
		}
		pages, err := otcCompute.List(computeClient, otcCompute.ListOpts{}).AllPages()
		if err != nil {
			return nil, err
		}
		servers, err := otcCompute.ExtractServers(pages)
		if err != nil {
			return nil, err
		}

		nameMap := buildECSNameMap(servers)
		statusFamilies := convertECSInstancesToMetrics(servers)
		aomFamilies := collectAOMMetrics(ctx, client, servers)

		return &EnrichResult{
			NameMap:       nameMap,
			ExtraFamilies: append(statusFamilies, aomFamilies...),
		}, nil
	})
}

// collectAOMMetrics fetches AOM host metrics for all servers and returns them
// as Prometheus MetricFamily objects with ecs_aom_ prefix.
//
// Fetches AOM metrics for all servers concurrently with bounded parallelism.
// Each server needs its own discovery call because per-device dimensions differ between servers.
func collectAOMMetrics(ctx context.Context, client *otcclient.Client, servers []otcCompute.Server) []*dto.MetricFamily {
	if len(servers) == 0 {
		return nil
	}

	aomClient, err := client.AOMV1()
	if err != nil {
		client.Logger.Warn("AOM client creation failed, skipping AOM metrics", "error", err.Error())
		return nil
	}

	// Fetch AOM metrics for all servers concurrently. Each server needs its
	// own discovery call because per-device dimensions (diskDevice, netDevice,
	// mountPoint) differ between servers.
	type serverResult struct {
		families []*dto.MetricFamily
		err      error
	}
	results := make([]serverResult, len(servers))

	sem := make(chan struct{}, Config.AOMConcurrency)
	var wg sync.WaitGroup

	for i, s := range servers {
		if ctx.Err() != nil {
			break
		}
		wg.Add(1)
		sem <- struct{}{}
		go func(idx int, server otcCompute.Server) {
			defer wg.Done()
			defer func() { <-sem }()

			resp, err := fetchAOMMetrics(ctx, aomClient, server.ID)
			if err != nil {
				results[idx] = serverResult{err: err}
				return
			}
			results[idx] = serverResult{
				families: convertAOMResponseToFamilies(resp, server.ID, server.Name),
			}
		}(i, s)
	}

	wg.Wait()

	var allFamilies []*dto.MetricFamily
	failCount := 0
	for i, r := range results {
		if r.err != nil {
			failCount++
			client.Logger.Warn("AOM fetch failed for server, skipping",
				"server_id", servers[i].ID, "server_name", servers[i].Name,
				"error", r.err.Error())
			continue
		}
		allFamilies = mergeMetricFamilies(allFamilies, r.families)
	}

	if failCount > 0 {
		client.Logger.Warn("AOM metric collection completed with failures",
			"total_servers", len(servers), "failed_servers", failCount)
	}
	if len(allFamilies) > 0 {
		client.Logger.Debug("AOM metrics collected", "families", len(allFamilies))
	}
	return allFamilies
}

// mergeMetricFamilies merges src families into dst. Families with the same name
// have their metrics combined into a single family.
func mergeMetricFamilies(dst, src []*dto.MetricFamily) []*dto.MetricFamily {
	index := make(map[string]*dto.MetricFamily)
	for _, f := range dst {
		index[f.GetName()] = f
	}
	for _, f := range src {
		if existing, ok := index[f.GetName()]; ok {
			existing.Metric = append(existing.Metric, f.Metric...)
		} else {
			index[f.GetName()] = f
			dst = append(dst, f)
		}
	}
	return dst
}

// buildECSNameMap creates a mapping from server ID to server name.
func buildECSNameMap(servers []otcCompute.Server) map[string]string {
	m := make(map[string]string, len(servers))
	for _, s := range servers {
		m[s.ID] = s.Name
	}
	return m
}

// convertECSInstancesToMetrics creates a MetricFamily "ecs_instance_status" with
// a gauge metric per server. The value is 0.0 for ACTIVE servers, 1.0 otherwise
// (OTC convention: 0=normal, 1=abnormal).
func convertECSInstancesToMetrics(servers []otcCompute.Server) []*dto.MetricFamily {
	metrics := make([]*dto.Metric, 0, len(servers))
	for _, s := range servers {
		value := 1.0
		if s.Status == "ACTIVE" {
			value = 0.0
		}
		metrics = append(metrics, NewGaugeMetric(value, map[string]string{
			"resource_id":   s.ID,
			"resource_name": s.Name,
			"status":        s.Status,
		}))
	}
	return []*dto.MetricFamily{NewGaugeMetricFamily("ecs_instance_status", metrics)}
}

