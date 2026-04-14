package provider

import (
	"context"

	dto "github.com/prometheus/client_model/go"

	"github.com/iits-consulting/otc-prometheus-exporter/otcclient"
)

// MetricProvider is the core abstraction for OTC service metric collection.
// Each OTC service (ECS, RDS, DMS, etc.) implements this interface.
type MetricProvider interface {
	// Namespace returns the OTC CES namespace, e.g. "SYS.ECS".
	Namespace() string
	// Collect fetches metrics from the OTC API and returns them as Prometheus MetricFamily objects.
	Collect(ctx context.Context, client *otcclient.Client) ([]*dto.MetricFamily, error)
}

// Registry maps namespace strings to MetricProvider implementations.
type Registry struct {
	providers map[string]MetricProvider
}

// NewRegistry creates an empty Registry.
func NewRegistry() *Registry {
	return &Registry{providers: make(map[string]MetricProvider)}
}

// Register adds a provider to the registry, keyed by its Namespace().
func (r *Registry) Register(p MetricProvider) {
	r.providers[p.Namespace()] = p
}

// Get returns the provider for the given namespace, if any.
func (r *Registry) Get(namespace string) (MetricProvider, bool) {
	p, ok := r.providers[namespace]
	return p, ok
}

// GetOrFallback returns the registered provider for the namespace.
// If no provider is registered, it returns a GenericCESProvider that
// forwards the namespace to the CES batch API.
func (r *Registry) GetOrFallback(namespace string) MetricProvider {
	if p, ok := r.providers[namespace]; ok {
		return p
	}
	return &GenericCESProvider{namespace: namespace}
}

// Namespaces returns all registered namespace strings.
func (r *Registry) Namespaces() []string {
	ns := make([]string, 0, len(r.providers))
	for k := range r.providers {
		ns = append(ns, k)
	}
	return ns
}

// All returns all registered MetricProvider implementations.
func (r *Registry) All() []MetricProvider {
	providers := make([]MetricProvider, 0, len(r.providers))
	for _, p := range r.providers {
		providers = append(providers, p)
	}
	return providers
}
