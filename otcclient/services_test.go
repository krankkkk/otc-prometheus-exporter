package otcclient

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
)

func newCache() clientCache {
	return clientCache{items: make(map[string]clientCacheEntry)}
}

func TestClientCacheReturnsCachedClient(t *testing.T) {
	cache := newCache()
	callCount := 0
	sc := &golangsdk.ServiceClient{Endpoint: "https://example.com"}

	factory := func() (*golangsdk.ServiceClient, error) {
		callCount++
		return sc, nil
	}

	first, err := cache.getOrCreate("test", factory)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	second, err := cache.getOrCreate("test", factory)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if first != second {
		t.Fatal("expected same client instance on second call")
	}
	if callCount != 1 {
		t.Fatalf("expected factory called once, got %d", callCount)
	}
}

func TestClientCacheDoesNotCacheErrors(t *testing.T) {
	cache := newCache()
	callCount := 0
	factoryErr := errors.New("endpoint discovery failed")

	factory := func() (*golangsdk.ServiceClient, error) {
		callCount++
		return nil, factoryErr
	}

	_, err := cache.getOrCreate("failing", factory)
	if !errors.Is(err, factoryErr) {
		t.Fatalf("expected factory error, got: %v", err)
	}

	// Second call should retry the factory since errors are not cached.
	_, err = cache.getOrCreate("failing", factory)
	if !errors.Is(err, factoryErr) {
		t.Fatalf("expected factory error on retry, got: %v", err)
	}
	if callCount != 2 {
		t.Fatalf("expected factory called twice (error not cached), got %d", callCount)
	}
}

func TestClientCacheIsolatesKeys(t *testing.T) {
	cache := newCache()
	scA := &golangsdk.ServiceClient{Endpoint: "https://a.example.com"}
	scB := &golangsdk.ServiceClient{Endpoint: "https://b.example.com"}

	a, err := cache.getOrCreate("a", func() (*golangsdk.ServiceClient, error) { return scA, nil })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b, err := cache.getOrCreate("b", func() (*golangsdk.ServiceClient, error) { return scB, nil })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if a.Endpoint == b.Endpoint {
		t.Fatal("expected different clients for different keys")
	}
}

func TestClientCacheConcurrentAccess(t *testing.T) {
	cache := newCache()
	var factoryCalls atomic.Int64
	sc := &golangsdk.ServiceClient{Endpoint: "https://example.com"}

	factory := func() (*golangsdk.ServiceClient, error) {
		factoryCalls.Add(1)
		return sc, nil
	}

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, err := cache.getOrCreate("shared", factory)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if got != sc {
				t.Errorf("expected same client instance")
			}
		}()
	}
	wg.Wait()

	if calls := factoryCalls.Load(); calls != 1 {
		t.Fatalf("expected factory called once under concurrency, got %d", calls)
	}
}

func TestRegionCESClientFallsBackWithoutRegionProvider(t *testing.T) {
	// Pre-populate the cache with a fake CES client so the fallback path
	// returns it without calling into the SDK (which needs a real provider).
	fakeCES := &golangsdk.ServiceClient{Endpoint: "https://ces.eu-de.otc.t-systems.com"}

	c := &Client{
		Logger: noopLogger{},
		cache: clientCache{items: map[string]clientCacheEntry{
			"ces": {client: fakeCES},
		}},
		Region: "eu-de",
		// regionProvider is nil — should fall back to CESClient
	}

	got, err := c.RegionCESClient()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != fakeCES {
		t.Fatal("expected RegionCESClient to fall back to cached CES client")
	}
}
