package otcclient

import (
	"context"
	"net/http"
)

// contextTransport injects a context into outgoing HTTP requests that have no
// context set (i.e. Done channel is nil). The gophertelekomcloud SDK creates
// requests via http.NewRequest() without context, so this transport adds the
// scrape context to enable cancellation of in-flight API calls.
//
// Requests that already carry a context (Done != nil) are left untouched.
type contextTransport struct {
	base http.RoundTripper
	ctx  context.Context
}

func (t *contextTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Context().Done() == nil && t.ctx != nil {
		req = req.WithContext(t.ctx)
	}
	return t.base.RoundTrip(req)
}

// WithContext returns a lightweight copy of the Client that injects ctx into
// all outgoing HTTP requests. The copy shares the underlying connection pool
// (http.Transport) and auth state with the original, but has its own
// contextTransport so concurrent scrapes each get the correct context.
//
// The returned Client should be used for a single Collect call and then
// discarded. ServiceClient caches are not shared — each WithContext client
// re-discovers endpoints from the in-memory catalog (no HTTP calls).
func (c *Client) WithContext(ctx context.Context) *Client {
	cp := &Client{
		cfg:             c.cfg,
		Region:          c.Region,
		Logger:          c.Logger,
		RegionProjectID: c.RegionProjectID,
		cache:           clientCache{items: make(map[string]clientCacheEntry)},
	}

	provCopy := *c.provider
	provCopy.HTTPClient = http.Client{
		Transport: &contextTransport{base: c.provider.HTTPClient.Transport, ctx: ctx},
		Timeout:   c.provider.HTTPClient.Timeout,
	}
	cp.provider = &provCopy

	if c.regionProvider != nil {
		regCopy := *c.regionProvider
		regCopy.HTTPClient = http.Client{
			Transport: &contextTransport{base: c.regionProvider.HTTPClient.Transport, ctx: ctx},
			Timeout:   c.regionProvider.HTTPClient.Timeout,
		}
		cp.regionProvider = &regCopy
	}

	return cp
}
