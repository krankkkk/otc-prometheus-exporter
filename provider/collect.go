package provider

import (
	"context"
	"fmt"
	"sync"

	dto "github.com/prometheus/client_model/go"

	"github.com/iits-consulting/otc-prometheus-exporter/otcclient"
)

// EnrichResult holds everything the service-API goroutine produced.
type EnrichResult struct {
	// NameMap is stored in the cache and used for EnrichWithNames (or fallback).
	NameMap map[string]string
	// ExtraFamilies are appended to CES families (e.g. status metrics).
	ExtraFamilies []*dto.MetricFamily
	// Enrich, if set, is called instead of EnrichWithNames. Use this for
	// providers that need custom enrichment logic (e.g. ELB's composite names).
	Enrich func([]*dto.MetricFamily)
}

// EnrichFunc fetches service-API data and returns an EnrichResult.
// Called in a goroutine parallel to CES collection.
type EnrichFunc func(ctx context.Context, client *otcclient.Client) (*EnrichResult, error)

// CollectWithEnrichment runs CES metrics collection and an enrichment function
// in parallel, handles cache fallback on enrichment failure, and returns the
// combined metric families.
func CollectWithEnrichment(
	ctx context.Context,
	client *otcclient.Client,
	namespace string,
	enrichFn EnrichFunc,
) ([]*dto.MetricFamily, error) {
	var (
		families  []*dto.MetricFamily
		cesErr    error
		result    *EnrichResult
		enrichErr error
	)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		if client == nil {
			return
		}
		families, cesErr = CollectCESMetrics(ctx, client, namespace)
		if cesErr != nil {
			client.Logger.Error("CES metrics failed", "namespace", namespace, "error", cesErr.Error())
		}
	}()

	go func() {
		defer wg.Done()
		if !ShouldEnrich(ctx) {
			return
		}
		result, enrichErr = enrichFn(ctx, client)
	}()

	wg.Wait()

	if !ShouldEnrich(ctx) {
		if len(families) == 0 {
			return nil, cesErr
		}
		return families, nil
	}

	if enrichErr != nil {
		if client != nil {
			client.Logger.Warn("enrichment failed", "namespace", namespace, "error", enrichErr.Error())
			if cached := Cache.Get(namespace); cached != nil {
				age := Cache.GetAge(namespace)
				client.Logger.Info("using cached names for enrichment",
					"namespace", namespace,
					"cached_entries", len(cached),
					"cache_age", fmt.Sprintf("%.0fm", age.Minutes()))
				EnrichWithNames(families, cached)
			}
		}
		if len(families) == 0 {
			return nil, cesErr
		}
		return families, nil
	}

	if client != nil {
		client.Logger.Debug("enrichment completed", "namespace", namespace)
	}

	Cache.Put(namespace, result.NameMap)
	if result.Enrich != nil {
		result.Enrich(families)
	} else {
		EnrichWithNames(families, result.NameMap)
	}

	families = append(families, result.ExtraFamilies...)
	if len(families) == 0 {
		return nil, cesErr
	}
	return families, nil
}
