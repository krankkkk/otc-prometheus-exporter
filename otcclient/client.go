package otcclient

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	golangsdk "github.com/opentelekomcloud/gophertelekomcloud"
	"github.com/opentelekomcloud/gophertelekomcloud/openstack"

	"github.com/iits-consulting/otc-prometheus-exporter/internal"
)

// Client wraps an authenticated OTC provider and exposes service client factories.
type Client struct {
	provider        *golangsdk.ProviderClient
	regionProvider  *golangsdk.ProviderClient // authenticated with region-level project (for global services like OBS)
	cfg             Config
	Region          string
	Logger          internal.ILogger
	RegionProjectID string // global/region-level project ID, discovered at startup
	cache           clientCache
}

// New creates an authenticated OTC client from the given configuration.
// It validates the config and returns an error if the config is invalid.
func New(cfg Config, logger internal.ILogger) (*Client, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	opts := authOptions(cfg)
	provider, err := openstack.AuthenticatedClient(opts)
	if err != nil {
		return nil, fmt.Errorf("otcclient: authentication failed: %w", err)
	}

	// Configure the HTTP transport for concurrent scraping. The default
	// transport allows only 2 idle connections per host, which causes TLS
	// handshake timeouts when Prometheus scrapes many namespaces in parallel.
	requestTimeout := cfg.RequestTimeout
	if requestTimeout == 0 {
		requestTimeout = 10 * time.Second
	}

	idleConnTimeout := cfg.IdleConnTimeout
	if idleConnTimeout == 0 {
		idleConnTimeout = 90 * time.Second
	}

	baseTransport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     &tls.Config{MinVersion: tls.VersionTLS12},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		MaxConnsPerHost:     10,
		IdleConnTimeout:     idleConnTimeout,
	}

	provider.HTTPClient = http.Client{
		Transport: newInstrumentedTransport(baseTransport, logger),
		Timeout:   requestTimeout,
	}

	return &Client{
		provider: provider,
		cfg:      cfg,
		Region:   cfg.Region,
		Logger:   logger,
		cache:    clientCache{items: make(map[string]clientCacheEntry)},
	}, nil
}

// endpointOpts returns the standard endpoint options used by all service client factories.
func (c *Client) endpointOpts() golangsdk.EndpointOpts {
	return golangsdk.EndpointOpts{Region: c.Region}
}
