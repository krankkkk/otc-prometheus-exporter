package provider

import (
	"context"
	"errors"
	"testing"

	dto "github.com/prometheus/client_model/go"

	"github.com/iits-consulting/otc-prometheus-exporter/otcclient"
)

// fakeCESFamilies builds a simple metric family for testing.
func fakeCESFamilies() []*dto.MetricFamily {
	m := NewGaugeMetric(1.0, map[string]string{
		"resource_id":   "res-1",
		"resource_name": "",
	})
	return []*dto.MetricFamily{NewGaugeMetricFamily("test_metric", []*dto.Metric{m})}
}

func TestCollectWithEnrichmentShouldEnrichFalse(t *testing.T) {
	ctx := WithEnrich(context.Background(), false)
	called := false

	families, err := CollectWithEnrichment(ctx, nil, "SYS.TEST", func(_ context.Context, _ *otcclient.Client) (*EnrichResult, error) {
		called = true
		return nil, nil
	})

	if called {
		t.Error("enrichFn should not be called when ShouldEnrich is false")
	}
	if len(families) > 0 {
		t.Error("expected nil/empty families with nil client")
	}
	_ = err
}

func TestCollectWithEnrichmentEnrichError_CacheFallback(t *testing.T) {
	Cache.Put("SYS.TESTFALLBACK", map[string]string{"res-1": "cached-name"})
	defer func() {
		Cache.Put("SYS.TESTFALLBACK", map[string]string{})
	}()

	enrichFn := func(_ context.Context, _ *otcclient.Client) (*EnrichResult, error) {
		return nil, errors.New("api timeout")
	}

	ctx := context.Background()
	_, _ = CollectWithEnrichment(ctx, nil, "SYS.TESTFALLBACK", enrichFn)
}

func TestEnrichResultCustomEnrich(t *testing.T) {
	families := fakeCESFamilies()
	customCalled := false

	result := &EnrichResult{
		NameMap: map[string]string{"res-1": "from-custom"},
		Enrich: func(fams []*dto.MetricFamily) {
			customCalled = true
			EnrichWithNames(fams, map[string]string{"res-1": "custom-name"})
		},
	}

	if result.Enrich != nil {
		result.Enrich(families)
	} else {
		EnrichWithNames(families, result.NameMap)
	}

	if !customCalled {
		t.Error("custom Enrich function was not called")
	}

	for _, lp := range families[0].Metric[0].Label {
		if lp.GetName() == "resource_name" && lp.GetValue() != "custom-name" {
			t.Errorf("expected resource_name 'custom-name', got %q", lp.GetValue())
		}
	}
}

func TestEnrichResultDefaultEnrich(t *testing.T) {
	families := fakeCESFamilies()

	result := &EnrichResult{
		NameMap: map[string]string{"res-1": "default-name"},
	}

	if result.Enrich != nil {
		result.Enrich(families)
	} else {
		EnrichWithNames(families, result.NameMap)
	}

	for _, lp := range families[0].Metric[0].Label {
		if lp.GetName() == "resource_name" && lp.GetValue() != "default-name" {
			t.Errorf("expected resource_name 'default-name', got %q", lp.GetValue())
		}
	}
}
