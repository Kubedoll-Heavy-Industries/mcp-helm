package helm

import (
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/hashicorp/golang-lru/v2/expirable"
	chartv2 "helm.sh/helm/v4/pkg/chart/v2"
	repo "helm.sh/helm/v4/pkg/repo/v1"
)

// CacheStats holds cache performance metrics.
type CacheStats struct {
	Hits   uint64 // Number of cache hits
	Misses uint64 // Number of cache misses
	Size   int    // Current number of entries
}

// IndexCache caches repository indexes with bounded size and TTL expiration.
// Thread-safe. Uses LRU eviction when capacity is reached.
type IndexCache struct {
	cache    *expirable.LRU[string, *repo.IndexFile]
	repoLock *repoLockManager
	hits     atomic.Uint64
	misses   atomic.Uint64
}

// NewIndexCache creates a bounded index cache.
// Entries expire after ttl and are evicted LRU when size exceeds capacity.
func NewIndexCache(capacity int, ttl time.Duration) *IndexCache {
	if capacity <= 0 {
		capacity = 100
	}
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &IndexCache{
		cache:    expirable.NewLRU[string, *repo.IndexFile](capacity, nil, ttl),
		repoLock: newRepoLockManager(),
	}
}

// Get retrieves a cached index if present and not expired.
func (c *IndexCache) Get(repoURL string) (*repo.IndexFile, bool) {
	index, ok := c.cache.Get(repoURL)
	if ok {
		c.hits.Add(1)
	} else {
		c.misses.Add(1)
	}
	return index, ok
}

// Stats returns cache performance metrics.
// All values are snapshotted into locals before constructing the struct
// to minimize the window for inconsistency between reads.
func (c *IndexCache) Stats() CacheStats {
	hits := c.hits.Load()
	misses := c.misses.Load()
	size := c.cache.Len()
	return CacheStats{
		Hits:   hits,
		Misses: misses,
		Size:   size,
	}
}

// Put stores an index in the cache.
func (c *IndexCache) Put(repoURL string, index *repo.IndexFile) {
	c.cache.Add(repoURL, index)
}

// Invalidate removes a specific repo from the cache.
func (c *IndexCache) Invalidate(repoURL string) {
	c.cache.Remove(repoURL)
}

// Clear removes all entries from the cache.
func (c *IndexCache) Clear() {
	c.cache.Purge()
}

// Len returns the current number of cached indexes.
func (c *IndexCache) Len() int {
	return c.cache.Len()
}

// LockRepo acquires a per-repo lock and returns an unlock function.
// Prevents thundering herd on cache miss.
func (c *IndexCache) LockRepo(repoURL string) func() {
	return c.repoLock.lock(repoURL)
}

// ChartCache caches loaded Helm charts with bounded size.
// Thread-safe. Uses LRU eviction when capacity is reached.
// No TTL since chart versions are immutable.
type ChartCache struct {
	cache  *lru.Cache[string, *chartv2.Chart]
	hits   atomic.Uint64
	misses atomic.Uint64
}

// NewChartCache creates a bounded chart cache.
// Charts don't expire (versions are immutable) but are evicted LRU when size exceeds capacity.
func NewChartCache(capacity int) *ChartCache {
	if capacity <= 0 {
		capacity = 50
	}
	cache, err := lru.New[string, *chartv2.Chart](capacity)
	if err != nil {
		// lru.New only fails if size <= 0, which we guard above.
		panic("helm: chart cache: " + err.Error())
	}
	return &ChartCache{
		cache: cache,
	}
}

// Get retrieves a chart from the cache.
func (c *ChartCache) Get(repoURL, chartName, version string) (*chartv2.Chart, bool) {
	chart, ok := c.cache.Get(makeChartKey(repoURL, chartName, version))
	if ok {
		c.hits.Add(1)
	} else {
		c.misses.Add(1)
	}
	return chart, ok
}

// Stats returns cache performance metrics.
// All values are snapshotted into locals before constructing the struct
// to minimize the window for inconsistency between reads.
func (c *ChartCache) Stats() CacheStats {
	hits := c.hits.Load()
	misses := c.misses.Load()
	size := c.cache.Len()
	return CacheStats{
		Hits:   hits,
		Misses: misses,
		Size:   size,
	}
}

// Put stores a chart in the cache.
func (c *ChartCache) Put(repoURL, chartName, version string, chart *chartv2.Chart) {
	c.cache.Add(makeChartKey(repoURL, chartName, version), chart)
}

// Clear removes all entries from the cache.
func (c *ChartCache) Clear() {
	c.cache.Purge()
}

// Len returns the current number of cached charts.
func (c *ChartCache) Len() int {
	return c.cache.Len()
}

// makeChartKey builds an unambiguous cache key using length-prefixed encoding.
// This prevents collisions when values contain the separator character.
func makeChartKey(repoURL, chartName, version string) string {
	return strconv.Itoa(len(repoURL)) + ":" + repoURL +
		strconv.Itoa(len(chartName)) + ":" + chartName +
		version
}

// repoLockManager manages per-repository locks for concurrent access.
// Prevents thundering herd when multiple goroutines try to fetch the same repo index.
type repoLockManager struct {
	mu    sync.Mutex
	locks map[string]*repoLock
}

type repoLock struct {
	mu       sync.Mutex
	refCount int
}

func newRepoLockManager() *repoLockManager {
	return &repoLockManager{
		locks: make(map[string]*repoLock),
	}
}

// lock acquires a lock for the specified repo and returns an unlock function.
// The lock is reference-counted and cleaned up when all holders release.
func (m *repoLockManager) lock(repoURL string) func() {
	m.mu.Lock()
	rl, ok := m.locks[repoURL]
	if !ok {
		rl = &repoLock{}
		m.locks[repoURL] = rl
	}
	rl.refCount++
	m.mu.Unlock()

	// If we fail to acquire the per-repo lock (e.g. panic), ensure
	// the refCount is decremented so the lock entry can be cleaned up.
	acquired := false
	defer func() {
		if !acquired {
			m.mu.Lock()
			rl.refCount--
			if rl.refCount == 0 {
				delete(m.locks, repoURL)
			}
			m.mu.Unlock()
		}
	}()

	rl.mu.Lock()
	acquired = true

	return func() {
		rl.mu.Unlock()

		m.mu.Lock()
		rl.refCount--
		if rl.refCount == 0 {
			delete(m.locks, repoURL)
		}
		m.mu.Unlock()
	}
}
