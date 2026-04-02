package provider

import (
	"testing"
)

func TestGenericCESProviderNamespace(t *testing.T) {
	p := &GenericCESProvider{namespace: "SYS.TEST"}
	if got := p.Namespace(); got != "SYS.TEST" {
		t.Errorf("expected namespace SYS.TEST, got %q", got)
	}
}

func TestGenericCESProviderCollectNilClient(t *testing.T) {
	p := &GenericCESProvider{namespace: "SYS.TEST"}
	families, err := p.Collect(t.Context(), nil)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if len(families) != 0 {
		t.Errorf("expected empty families, got %d", len(families))
	}
}

func TestRegistryGetOrFallbackRegistered(t *testing.T) {
	r := NewRegistry()
	p := &GenericCESProvider{namespace: "SYS.ECS"}
	r.Register(p)

	got := r.GetOrFallback("SYS.ECS")
	if got != p {
		t.Errorf("expected registered provider, got %T", got)
	}
}

func TestRegistryGetOrFallbackUnregistered(t *testing.T) {
	r := NewRegistry()

	got := r.GetOrFallback("SYS.UNKNOWN")
	generic, ok := got.(*GenericCESProvider)
	if !ok {
		t.Fatalf("expected *GenericCESProvider, got %T", got)
	}
	if generic.namespace != "SYS.UNKNOWN" {
		t.Errorf("expected namespace SYS.UNKNOWN, got %q", generic.namespace)
	}
}
