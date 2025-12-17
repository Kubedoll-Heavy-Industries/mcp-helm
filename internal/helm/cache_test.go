package helm

import (
	"testing"
	"time"

	chartv2 "helm.sh/helm/v4/pkg/chart/v2"
	repo "helm.sh/helm/v4/pkg/repo/v1"
)

func TestIndexCache(t *testing.T) {
	t.Run("get missing returns false", func(t *testing.T) {
		cache := NewIndexCache(time.Minute)
		_, ok := cache.Get("https://example.com")
		if ok {
			t.Error("expected Get to return false for missing key")
		}
	})

	t.Run("put and get", func(t *testing.T) {
		cache := NewIndexCache(time.Minute)
		index := &repo.IndexFile{
			Entries: map[string]repo.ChartVersions{
				"nginx": {{Metadata: &chartv2.Metadata{Name: "nginx"}}},
			},
		}

		cache.Put("https://example.com", index)

		got, ok := cache.Get("https://example.com")
		if !ok {
			t.Fatal("expected Get to return true")
		}
		if got != index {
			t.Error("expected same index back")
		}
	})

	t.Run("expired entry returns false", func(t *testing.T) {
		cache := NewIndexCache(time.Millisecond)
		index := &repo.IndexFile{}

		cache.Put("https://example.com", index)
		time.Sleep(5 * time.Millisecond)

		_, ok := cache.Get("https://example.com")
		if ok {
			t.Error("expected expired entry to return false")
		}
	})

	t.Run("invalidate removes entry", func(t *testing.T) {
		cache := NewIndexCache(time.Minute)
		cache.Put("https://example.com", &repo.IndexFile{})

		cache.Invalidate("https://example.com")

		_, ok := cache.Get("https://example.com")
		if ok {
			t.Error("expected invalidated entry to be gone")
		}
	})

	t.Run("clear removes all entries", func(t *testing.T) {
		cache := NewIndexCache(time.Minute)
		cache.Put("https://a.com", &repo.IndexFile{})
		cache.Put("https://b.com", &repo.IndexFile{})

		cache.Clear()

		if _, ok := cache.Get("https://a.com"); ok {
			t.Error("expected a.com to be cleared")
		}
		if _, ok := cache.Get("https://b.com"); ok {
			t.Error("expected b.com to be cleared")
		}
	})

	t.Run("default TTL on zero", func(t *testing.T) {
		cache := NewIndexCache(0)
		if cache.ttl != 5*time.Minute {
			t.Errorf("expected default TTL of 5m, got %v", cache.ttl)
		}
	})
}

func TestChartCache(t *testing.T) {
	makeChart := func(name string) *chartv2.Chart {
		return &chartv2.Chart{
			Metadata: &chartv2.Metadata{Name: name, Version: "1.0.0"},
		}
	}

	t.Run("get missing returns false", func(t *testing.T) {
		cache := NewChartCache(10)
		_, ok := cache.Get("repo", "chart", "1.0.0")
		if ok {
			t.Error("expected Get to return false for missing key")
		}
	})

	t.Run("put and get", func(t *testing.T) {
		cache := NewChartCache(10)
		chart := makeChart("nginx")

		cache.Put("https://repo.com", "nginx", "1.0.0", chart)

		got, ok := cache.Get("https://repo.com", "nginx", "1.0.0")
		if !ok {
			t.Fatal("expected Get to return true")
		}
		if got != chart {
			t.Error("expected same chart back")
		}
	})

	t.Run("different versions are separate", func(t *testing.T) {
		cache := NewChartCache(10)
		v1 := makeChart("nginx-v1")
		v2 := makeChart("nginx-v2")

		cache.Put("repo", "nginx", "1.0.0", v1)
		cache.Put("repo", "nginx", "2.0.0", v2)

		got1, _ := cache.Get("repo", "nginx", "1.0.0")
		got2, _ := cache.Get("repo", "nginx", "2.0.0")

		if got1.Metadata.Name != "nginx-v1" {
			t.Error("v1 was overwritten")
		}
		if got2.Metadata.Name != "nginx-v2" {
			t.Error("v2 not stored correctly")
		}
	})

	t.Run("LRU eviction", func(t *testing.T) {
		cache := NewChartCache(2)

		cache.Put("repo", "a", "1.0.0", makeChart("a"))
		cache.Put("repo", "b", "1.0.0", makeChart("b"))
		cache.Put("repo", "c", "1.0.0", makeChart("c")) // Should evict "a"

		if _, ok := cache.Get("repo", "a", "1.0.0"); ok {
			t.Error("expected 'a' to be evicted")
		}
		if _, ok := cache.Get("repo", "b", "1.0.0"); !ok {
			t.Error("expected 'b' to still be present")
		}
		if _, ok := cache.Get("repo", "c", "1.0.0"); !ok {
			t.Error("expected 'c' to be present")
		}
	})

	t.Run("access updates LRU order", func(t *testing.T) {
		cache := NewChartCache(2)

		cache.Put("repo", "a", "1.0.0", makeChart("a"))
		cache.Put("repo", "b", "1.0.0", makeChart("b"))

		// Access "a" to make it most recently used
		cache.Get("repo", "a", "1.0.0")

		// Add "c" - should evict "b" (least recently used)
		cache.Put("repo", "c", "1.0.0", makeChart("c"))

		if _, ok := cache.Get("repo", "a", "1.0.0"); !ok {
			t.Error("expected 'a' to still be present after access")
		}
		if _, ok := cache.Get("repo", "b", "1.0.0"); ok {
			t.Error("expected 'b' to be evicted")
		}
	})

	t.Run("update existing entry", func(t *testing.T) {
		cache := NewChartCache(10)
		v1 := makeChart("nginx-v1")
		v2 := makeChart("nginx-v2")

		cache.Put("repo", "nginx", "1.0.0", v1)
		cache.Put("repo", "nginx", "1.0.0", v2)

		got, _ := cache.Get("repo", "nginx", "1.0.0")
		if got.Metadata.Name != "nginx-v2" {
			t.Error("expected updated chart")
		}
		if cache.Size() != 1 {
			t.Errorf("expected size 1, got %d", cache.Size())
		}
	})

	t.Run("size returns correct count", func(t *testing.T) {
		cache := NewChartCache(10)
		if cache.Size() != 0 {
			t.Error("expected empty cache to have size 0")
		}

		cache.Put("repo", "a", "1.0.0", makeChart("a"))
		cache.Put("repo", "b", "1.0.0", makeChart("b"))

		if cache.Size() != 2 {
			t.Errorf("expected size 2, got %d", cache.Size())
		}
	})

	t.Run("clear empties cache", func(t *testing.T) {
		cache := NewChartCache(10)
		cache.Put("repo", "a", "1.0.0", makeChart("a"))
		cache.Put("repo", "b", "1.0.0", makeChart("b"))

		cache.Clear()

		if cache.Size() != 0 {
			t.Errorf("expected size 0 after clear, got %d", cache.Size())
		}
	})

	t.Run("default capacity on zero", func(t *testing.T) {
		cache := NewChartCache(0)
		if cache.capacity != 50 {
			t.Errorf("expected default capacity of 50, got %d", cache.capacity)
		}
	})
}

func TestRepoLockManager(t *testing.T) {
	t.Run("lock and unlock", func(t *testing.T) {
		m := newRepoLockManager()

		unlock := m.lock("https://repo.com")
		unlock()
		// Should not deadlock
	})

	t.Run("different repos can lock concurrently", func(t *testing.T) {
		m := newRepoLockManager()

		done := make(chan bool, 2)

		go func() {
			unlock := m.lock("https://repo-a.com")
			time.Sleep(10 * time.Millisecond)
			unlock()
			done <- true
		}()

		go func() {
			unlock := m.lock("https://repo-b.com")
			time.Sleep(10 * time.Millisecond)
			unlock()
			done <- true
		}()

		// Both should complete quickly
		timeout := time.After(100 * time.Millisecond)
		for i := 0; i < 2; i++ {
			select {
			case <-done:
			case <-timeout:
				t.Fatal("concurrent locks on different repos should not block")
			}
		}
	})

	t.Run("same repo locks are serialized", func(t *testing.T) {
		m := newRepoLockManager()

		var order []int
		done := make(chan bool)

		unlock1 := m.lock("https://repo.com")

		go func() {
			unlock2 := m.lock("https://repo.com")
			order = append(order, 2)
			unlock2()
			done <- true
		}()

		// Give goroutine time to block on lock
		time.Sleep(10 * time.Millisecond)
		order = append(order, 1)
		unlock1()

		<-done

		if len(order) != 2 || order[0] != 1 || order[1] != 2 {
			t.Errorf("expected order [1, 2], got %v", order)
		}
	})

	t.Run("cleanup after all unlocks", func(t *testing.T) {
		m := newRepoLockManager()

		unlock1 := m.lock("https://repo.com")
		unlock1()

		m.mu.Lock()
		count := len(m.locks)
		m.mu.Unlock()

		if count != 0 {
			t.Errorf("expected locks map to be empty after unlock, got %d entries", count)
		}
	})
}
