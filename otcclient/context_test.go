package otcclient

import (
	"context"
	"net/http"
	"testing"

	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
)

// roundTripFunc adapts a function to http.RoundTripper for testing.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestContextTransportInjectsContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var gotCtx context.Context
	base := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		gotCtx = req.Context()
		return &http.Response{StatusCode: 200}, nil
	})

	transport := &contextTransport{base: base, ctx: ctx}
	// SDK creates requests with http.NewRequest (context.Background → Done==nil).
	req, _ := http.NewRequest("GET", "https://example.com", nil)
	_, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotCtx != ctx {
		t.Fatal("expected transport to inject context into request with no context set")
	}
}

func TestContextTransportPreservesExistingContext(t *testing.T) {
	transportCtx := context.WithValue(context.Background(), "transport", true)
	requestCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var gotCtx context.Context
	base := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		gotCtx = req.Context()
		return &http.Response{StatusCode: 200}, nil
	})

	transport := &contextTransport{base: base, ctx: transportCtx}
	req, _ := http.NewRequestWithContext(requestCtx, "GET", "https://example.com", nil)
	_, _ = transport.RoundTrip(req)

	if gotCtx != requestCtx {
		t.Fatal("transport should preserve the request's existing context")
	}
}

func TestWithContextReturnsNewClient(t *testing.T) {
	original := &Client{
		provider: &golangsdk.ProviderClient{
			HTTPClient: http.Client{
				Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					return &http.Response{StatusCode: 200}, nil
				}),
			},
		},
		Region: "eu-de",
		Logger: noopLogger{},
		cache:  clientCache{items: make(map[string]clientCacheEntry)},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	derived := original.WithContext(ctx)

	if derived == original {
		t.Fatal("WithContext should return a new Client, not the same pointer")
	}
	if derived.provider == original.provider {
		t.Fatal("WithContext should shallow-copy the provider")
	}
	if derived.Region != original.Region {
		t.Fatal("WithContext should preserve Region")
	}
	if derived.Logger != original.Logger {
		t.Fatal("WithContext should share the same Logger")
	}
}

func TestWithContextInjectsContextIntoSDKRequests(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var gotCtx context.Context
	base := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		gotCtx = req.Context()
		return &http.Response{StatusCode: 200}, nil
	})

	original := &Client{
		provider: &golangsdk.ProviderClient{
			HTTPClient: http.Client{Transport: base},
		},
		Region: "eu-de",
		Logger: noopLogger{},
		cache:  clientCache{items: make(map[string]clientCacheEntry)},
	}

	derived := original.WithContext(ctx)

	// Simulate an SDK-style request (no context on request).
	req, _ := http.NewRequest("GET", "https://example.com", nil)
	_, _ = derived.provider.HTTPClient.Do(req)

	if gotCtx != ctx {
		t.Fatal("expected derived client's transport to inject the scrape context")
	}
}

func TestWithContextCopiesRegionProvider(t *testing.T) {
	base := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200}, nil
	})

	original := &Client{
		provider: &golangsdk.ProviderClient{
			HTTPClient: http.Client{Transport: base},
		},
		regionProvider: &golangsdk.ProviderClient{
			HTTPClient: http.Client{Transport: base},
		},
		Region: "eu-de",
		Logger: noopLogger{},
		cache:  clientCache{items: make(map[string]clientCacheEntry)},
	}

	ctx := context.Background()
	derived := original.WithContext(ctx)

	if derived.regionProvider == nil {
		t.Fatal("expected regionProvider to be copied")
	}
	if derived.regionProvider == original.regionProvider {
		t.Fatal("expected regionProvider to be a new pointer")
	}
}
