package provider

import (
	"sync"
	"testing"
	"time"
)

func TestNameCacheGetReturnsNilForUnknownNamespace(t *testing.T) {
	c := NewNameCache()
	if got := c.Get("SYS.RDS"); got != nil {
		t.Errorf("expected nil for unknown namespace, got %v", got)
	}
}

func TestNameCachePutAndGet(t *testing.T) {
	c := NewNameCache()
	names := map[string]string{"id-1": "server-1", "id-2": "server-2"}
	c.Put("SYS.ECS", names)

	got := c.Get("SYS.ECS")
	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
	if got["id-1"] != "server-1" {
		t.Errorf("expected id-1 -> server-1, got %q", got["id-1"])
	}
	if got["id-2"] != "server-2" {
		t.Errorf("expected id-2 -> server-2, got %q", got["id-2"])
	}
}

func TestNameCachePutOverwritesPreviousEntry(t *testing.T) {
	c := NewNameCache()
	c.Put("SYS.RDS", map[string]string{"id-1": "old-name"})
	c.Put("SYS.RDS", map[string]string{"id-1": "new-name", "id-2": "added"})

	got := c.Get("SYS.RDS")
	if len(got) != 2 {
		t.Fatalf("expected 2 entries after overwrite, got %d", len(got))
	}
	if got["id-1"] != "new-name" {
		t.Errorf("expected id-1 -> new-name, got %q", got["id-1"])
	}
}

func TestNameCachePutNilIsNoOp(t *testing.T) {
	c := NewNameCache()
	c.Put("SYS.RDS", nil)
	if got := c.Get("SYS.RDS"); got != nil {
		t.Errorf("expected nil after putting nil, got %v", got)
	}
}

func TestNameCacheGetReturnsDefensiveCopy(t *testing.T) {
	c := NewNameCache()
	c.Put("SYS.ECS", map[string]string{"id-1": "original"})

	got := c.Get("SYS.ECS")
	got["id-1"] = "mutated"
	got["id-2"] = "injected"

	original := c.Get("SYS.ECS")
	if original["id-1"] != "original" {
		t.Errorf("mutation leaked into cache: expected 'original', got %q", original["id-1"])
	}
	if len(original) != 1 {
		t.Errorf("injection leaked into cache: expected 1 entry, got %d", len(original))
	}
}

func TestNameCacheGetAgeReturnsZeroForUnknown(t *testing.T) {
	c := NewNameCache()
	if age := c.GetAge("SYS.RDS"); age != 0 {
		t.Errorf("expected 0 age for unknown namespace, got %v", age)
	}
}

func TestNameCacheGetAgeReturnsDuration(t *testing.T) {
	c := NewNameCache()
	c.Put("SYS.RDS", map[string]string{"id": "name"})
	time.Sleep(5 * time.Millisecond)

	age := c.GetAge("SYS.RDS")
	if age < 5*time.Millisecond {
		t.Errorf("expected age >= 5ms, got %v", age)
	}
}

func TestNameCacheConcurrentAccess(t *testing.T) {
	c := NewNameCache()
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			c.Put("SYS.ECS", map[string]string{"id": "name"})
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			_ = c.Get("SYS.ECS")
			_ = c.GetAge("SYS.ECS")
		}
	}()

	wg.Wait()
}
