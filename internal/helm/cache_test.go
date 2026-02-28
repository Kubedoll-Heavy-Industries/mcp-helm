package helm

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	chartv2 "helm.sh/helm/v4/pkg/chart/v2"
	repo "helm.sh/helm/v4/pkg/repo/v1"
)

func TestIndexCache(t *testing.T) {
	t.Run("get missing returns false", func(t *testing.T) {
		cache := NewIndexCache(10, time.Minute)
		_, ok := cache.Get("https://example.com")
		assert.False(t, ok)
	})

	t.Run("put and get", func(t *testing.T) {
		cache := NewIndexCache(10, time.Minute)
		index := &repo.IndexFile{
			Entries: map[string]repo.ChartVersions{
				"nginx": {{Metadata: &chartv2.Metadata{Name: "nginx"}}},
			},
		}

		cache.Put("https://example.com", index)

		got, ok := cache.Get("https://example.com")
		require.True(t, ok)
		assert.Equal(t, index, got)
	})

	t.Run("expired entry returns false", func(t *testing.T) {
		cache := NewIndexCache(10, 10*time.Millisecond)
		cache.Put("https://example.com", &repo.IndexFile{})

		time.Sleep(50 * time.Millisecond)

		_, ok := cache.Get("https://example.com")
		assert.False(t, ok, "expired entry should not be returned")
	})

	t.Run("invalidate removes entry", func(t *testing.T) {
		cache := NewIndexCache(10, time.Minute)
		cache.Put("https://example.com", &repo.IndexFile{})

		cache.Invalidate("https://example.com")

		_, ok := cache.Get("https://example.com")
		assert.False(t, ok)
	})

	t.Run("clear removes all entries", func(t *testing.T) {
		cache := NewIndexCache(10, time.Minute)
		cache.Put("https://a.com", &repo.IndexFile{})
		cache.Put("https://b.com", &repo.IndexFile{})

		cache.Clear()

		_, okA := cache.Get("https://a.com")
		_, okB := cache.Get("https://b.com")
		assert.False(t, okA)
		assert.False(t, okB)
	})

	t.Run("LRU eviction when capacity exceeded", func(t *testing.T) {
		cache := NewIndexCache(2, time.Minute)

		cache.Put("https://a.com", &repo.IndexFile{})
		cache.Put("https://b.com", &repo.IndexFile{})
		cache.Put("https://c.com", &repo.IndexFile{}) // Evicts a.com

		_, okA := cache.Get("https://a.com")
		_, okB := cache.Get("https://b.com")
		_, okC := cache.Get("https://c.com")

		assert.False(t, okA, "oldest entry should be evicted")
		assert.True(t, okB)
		assert.True(t, okC)
	})

	t.Run("Len returns correct count", func(t *testing.T) {
		cache := NewIndexCache(10, time.Minute)
		assert.Equal(t, 0, cache.Len())

		cache.Put("https://a.com", &repo.IndexFile{})
		cache.Put("https://b.com", &repo.IndexFile{})
		assert.Equal(t, 2, cache.Len())
	})

	t.Run("default values on zero", func(t *testing.T) {
		cache := NewIndexCache(0, 0)
		// Should not panic, uses defaults
		cache.Put("https://example.com", &repo.IndexFile{})
		_, ok := cache.Get("https://example.com")
		assert.True(t, ok)
	})

	t.Run("Stats tracks hits and misses", func(t *testing.T) {
		cache := NewIndexCache(10, time.Minute)

		// Initial stats should be zero
		stats := cache.Stats()
		assert.Equal(t, uint64(0), stats.Hits)
		assert.Equal(t, uint64(0), stats.Misses)
		assert.Equal(t, 0, stats.Size)

		// Miss
		cache.Get("https://missing.com")
		stats = cache.Stats()
		assert.Equal(t, uint64(0), stats.Hits)
		assert.Equal(t, uint64(1), stats.Misses)

		// Put and hit
		cache.Put("https://example.com", &repo.IndexFile{})
		cache.Get("https://example.com")
		stats = cache.Stats()
		assert.Equal(t, uint64(1), stats.Hits)
		assert.Equal(t, uint64(1), stats.Misses)
		assert.Equal(t, 1, stats.Size)
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
		assert.False(t, ok)
	})

	t.Run("put and get", func(t *testing.T) {
		cache := NewChartCache(10)
		chart := makeChart("nginx")

		cache.Put("https://repo.com", "nginx", "1.0.0", chart)

		got, ok := cache.Get("https://repo.com", "nginx", "1.0.0")
		require.True(t, ok)
		assert.Equal(t, chart, got)
	})

	t.Run("different versions are separate", func(t *testing.T) {
		cache := NewChartCache(10)
		v1 := makeChart("nginx-v1")
		v2 := makeChart("nginx-v2")

		cache.Put("repo", "nginx", "1.0.0", v1)
		cache.Put("repo", "nginx", "2.0.0", v2)

		got1, _ := cache.Get("repo", "nginx", "1.0.0")
		got2, _ := cache.Get("repo", "nginx", "2.0.0")

		assert.Equal(t, "nginx-v1", got1.Metadata.Name)
		assert.Equal(t, "nginx-v2", got2.Metadata.Name)
	})

	t.Run("LRU eviction", func(t *testing.T) {
		cache := NewChartCache(2)

		cache.Put("repo", "a", "1.0.0", makeChart("a"))
		cache.Put("repo", "b", "1.0.0", makeChart("b"))
		cache.Put("repo", "c", "1.0.0", makeChart("c")) // Evicts a

		_, okA := cache.Get("repo", "a", "1.0.0")
		_, okB := cache.Get("repo", "b", "1.0.0")
		_, okC := cache.Get("repo", "c", "1.0.0")

		assert.False(t, okA, "oldest entry should be evicted")
		assert.True(t, okB)
		assert.True(t, okC)
	})

	t.Run("access updates LRU order", func(t *testing.T) {
		cache := NewChartCache(2)

		cache.Put("repo", "a", "1.0.0", makeChart("a"))
		cache.Put("repo", "b", "1.0.0", makeChart("b"))

		// Access a to make it most recently used
		cache.Get("repo", "a", "1.0.0")

		// Add c - should evict b (least recently used)
		cache.Put("repo", "c", "1.0.0", makeChart("c"))

		_, okA := cache.Get("repo", "a", "1.0.0")
		_, okB := cache.Get("repo", "b", "1.0.0")

		assert.True(t, okA, "recently accessed should not be evicted")
		assert.False(t, okB, "least recently used should be evicted")
	})

	t.Run("Len returns correct count", func(t *testing.T) {
		cache := NewChartCache(10)
		assert.Equal(t, 0, cache.Len())

		cache.Put("repo", "a", "1.0.0", makeChart("a"))
		cache.Put("repo", "b", "1.0.0", makeChart("b"))
		assert.Equal(t, 2, cache.Len())
	})

	t.Run("clear empties cache", func(t *testing.T) {
		cache := NewChartCache(10)
		cache.Put("repo", "a", "1.0.0", makeChart("a"))

		cache.Clear()

		assert.Equal(t, 0, cache.Len())
	})

	t.Run("default capacity on zero", func(t *testing.T) {
		cache := NewChartCache(0)
		// Should not panic, uses default
		cache.Put("repo", "a", "1.0.0", makeChart("a"))
		_, ok := cache.Get("repo", "a", "1.0.0")
		assert.True(t, ok)
	})

	t.Run("Stats tracks hits and misses", func(t *testing.T) {
		cache := NewChartCache(10)

		// Initial stats should be zero
		stats := cache.Stats()
		assert.Equal(t, uint64(0), stats.Hits)
		assert.Equal(t, uint64(0), stats.Misses)
		assert.Equal(t, 0, stats.Size)

		// Miss
		cache.Get("repo", "missing", "1.0.0")
		stats = cache.Stats()
		assert.Equal(t, uint64(0), stats.Hits)
		assert.Equal(t, uint64(1), stats.Misses)

		// Put and hit
		cache.Put("repo", "nginx", "1.0.0", makeChart("nginx"))
		cache.Get("repo", "nginx", "1.0.0")
		stats = cache.Stats()
		assert.Equal(t, uint64(1), stats.Hits)
		assert.Equal(t, uint64(1), stats.Misses)
		assert.Equal(t, 1, stats.Size)
	})
}

func TestIndexCache_Stats_Concurrent(t *testing.T) {
	cache := NewIndexCache(100, time.Minute)

	// Pre-populate some entries
	for i := 0; i < 10; i++ {
		cache.Put("https://repo-"+strconv.Itoa(i)+".com", &repo.IndexFile{})
	}

	var wg sync.WaitGroup
	const goroutines = 20
	const iterations = 500

	// Goroutines performing gets (mix of hits and misses)
	for i := 0; i < goroutines/2; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				cache.Get("https://repo-" + strconv.Itoa(j%15) + ".com")
			}
		}(i)
	}

	// Goroutines reading Stats concurrently
	for i := 0; i < goroutines/2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				stats := cache.Stats()
				// Hits and misses should never go backwards individually
				_ = stats.Hits
				_ = stats.Misses
				assert.GreaterOrEqual(t, stats.Size, 0)
			}
		}()
	}

	wg.Wait()

	// Final stats should reflect all operations
	stats := cache.Stats()
	assert.Greater(t, stats.Hits+stats.Misses, uint64(0), "should have recorded operations")
}

func TestChartCache_Stats_Concurrent(t *testing.T) {
	cache := NewChartCache(100)

	makeChart := func(name string) *chartv2.Chart {
		return &chartv2.Chart{
			Metadata: &chartv2.Metadata{Name: name, Version: "1.0.0"},
		}
	}

	// Pre-populate some entries
	for i := 0; i < 10; i++ {
		name := "chart-" + strconv.Itoa(i)
		cache.Put("repo", name, "1.0.0", makeChart(name))
	}

	var wg sync.WaitGroup
	const goroutines = 20
	const iterations = 500

	// Goroutines performing gets (mix of hits and misses)
	for i := 0; i < goroutines/2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				cache.Get("repo", "chart-"+strconv.Itoa(j%15), "1.0.0")
			}
		}()
	}

	// Goroutines reading Stats concurrently
	for i := 0; i < goroutines/2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				stats := cache.Stats()
				_ = stats.Hits
				_ = stats.Misses
				assert.GreaterOrEqual(t, stats.Size, 0)
			}
		}()
	}

	wg.Wait()

	stats := cache.Stats()
	assert.Greater(t, stats.Hits+stats.Misses, uint64(0), "should have recorded operations")
}

func TestMakeChartKey_NoCollision(t *testing.T) {
	// These inputs previously collided when using \x00 as separator.
	// With length-prefixed encoding, they must produce distinct keys.
	tests := []struct {
		name string
		a    [3]string // repoURL, chartName, version
		b    [3]string
	}{
		{
			name: "null byte in chart name",
			a:    [3]string{"repo", "chart\x00", "1.0.0"},
			b:    [3]string{"repo", "chart", "\x001.0.0"},
		},
		{
			name: "name/version boundary shift",
			a:    [3]string{"repo", "abc", "def"},
			b:    [3]string{"repo", "ab", "cdef"},
		},
		{
			name: "repo/name boundary shift",
			a:    [3]string{"repo-a", "chart", "1.0.0"},
			b:    [3]string{"repo", "a-chart", "1.0.0"},
		},
		{
			name: "empty chart name vs non-empty",
			a:    [3]string{"repo", "", "v1"},
			b:    [3]string{"repo", "v", "1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyA := makeChartKey(tt.a[0], tt.a[1], tt.a[2])
			keyB := makeChartKey(tt.b[0], tt.b[1], tt.b[2])
			assert.NotEqual(t, keyA, keyB, "keys must not collide: %q vs %q", keyA, keyB)
		})
	}
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

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			unlock := m.lock("https://repo-a.com")
			time.Sleep(10 * time.Millisecond)
			unlock()
		}()

		go func() {
			defer wg.Done()
			unlock := m.lock("https://repo-b.com")
			time.Sleep(10 * time.Millisecond)
			unlock()
		}()

		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Success
		case <-time.After(100 * time.Millisecond):
			t.Fatal("concurrent locks on different repos should not block")
		}
	})

	t.Run("same repo locks are serialized", func(t *testing.T) {
		m := newRepoLockManager()

		var mu sync.Mutex
		var order []int

		unlock1 := m.lock("https://repo.com")

		done := make(chan struct{})
		go func() {
			unlock2 := m.lock("https://repo.com")
			mu.Lock()
			order = append(order, 2)
			mu.Unlock()
			unlock2()
			close(done)
		}()

		// Give goroutine time to block on lock
		time.Sleep(10 * time.Millisecond)
		mu.Lock()
		order = append(order, 1)
		mu.Unlock()
		unlock1()

		<-done

		assert.Equal(t, []int{1, 2}, order)
	})

	t.Run("cleanup after all unlocks", func(t *testing.T) {
		m := newRepoLockManager()

		unlock := m.lock("https://repo.com")
		unlock()

		m.mu.Lock()
		count := len(m.locks)
		m.mu.Unlock()

		assert.Equal(t, 0, count, "locks map should be empty after unlock")
	})

	t.Run("cleanup after concurrent unlocks", func(t *testing.T) {
		m := newRepoLockManager()
		const repoURL = "https://repo.com"

		// Acquire the first lock and hold it while others queue
		unlock1 := m.lock(repoURL)

		var wg sync.WaitGroup
		wg.Add(5)
		for i := 0; i < 5; i++ {
			go func() {
				defer wg.Done()
				u := m.lock(repoURL)
				u()
			}()
		}

		// Let goroutines queue up on the per-repo lock
		time.Sleep(20 * time.Millisecond)
		unlock1()

		wg.Wait()

		m.mu.Lock()
		count := len(m.locks)
		m.mu.Unlock()

		assert.Equal(t, 0, count, "locks map should be empty after all concurrent unlocks")
	})
}
