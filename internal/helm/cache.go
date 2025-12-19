package helm

import (
	"container/list"
	"sync"
	"time"

	chartv2 "helm.sh/helm/v4/pkg/chart/v2"
	"helm.sh/helm/v4/pkg/repo/v1"
)

// IndexCache caches repository indexes with TTL-based expiration.
type IndexCache struct {
	mu       sync.RWMutex
	ttl      time.Duration
	entries  map[string]*indexEntry
	repoLock *repoLockManager
}

type indexEntry struct {
	index     *repo.IndexFile
	fetchedAt time.Time
}

// NewIndexCache creates a new index cache with the specified TTL.
func NewIndexCache(ttl time.Duration) *IndexCache {
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &IndexCache{
		ttl:      ttl,
		entries:  make(map[string]*indexEntry),
		repoLock: newRepoLockManager(),
	}
}

// Get retrieves a cached index if it exists and is not expired.
func (c *IndexCache) Get(repoURL string) (*repo.IndexFile, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.entries[repoURL]
	if !ok {
		return nil, false
	}
	if time.Since(entry.fetchedAt) > c.ttl {
		return nil, false
	}
	return entry.index, true
}

// Put stores an index in the cache.
func (c *IndexCache) Put(repoURL string, index *repo.IndexFile) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[repoURL] = &indexEntry{
		index:     index,
		fetchedAt: time.Now(),
	}
}

// Invalidate removes a specific repo from the cache.
func (c *IndexCache) Invalidate(repoURL string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, repoURL)
}

// Clear removes all entries from the cache.
func (c *IndexCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*indexEntry)
}

// LockRepo acquires a per-repo lock and returns an unlock function.
// Use with defer: defer cache.LockRepo(url)()
func (c *IndexCache) LockRepo(repoURL string) func() {
	return c.repoLock.lock(repoURL)
}

// ChartCache implements an LRU cache for loaded Helm charts.
type ChartCache struct {
	mu       sync.Mutex
	capacity int
	cache    map[string]*list.Element
	lru      *list.List
}

type chartEntry struct {
	key   string
	chart *chartv2.Chart
}

// NewChartCache creates a new chart cache with the specified capacity.
func NewChartCache(capacity int) *ChartCache {
	if capacity <= 0 {
		capacity = 50
	}
	return &ChartCache{
		capacity: capacity,
		cache:    make(map[string]*list.Element),
		lru:      list.New(),
	}
}

// Get retrieves a chart from the cache.
func (c *ChartCache) Get(repoURL, chartName, version string) (*chartv2.Chart, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := makeChartKey(repoURL, chartName, version)
	elem, ok := c.cache[key]
	if !ok {
		return nil, false
	}

	// Move to front (most recently used)
	c.lru.MoveToFront(elem)
	return elem.Value.(*chartEntry).chart, true
}

// Put stores a chart in the cache, evicting the oldest entry if necessary.
func (c *ChartCache) Put(repoURL, chartName, version string, chart *chartv2.Chart) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := makeChartKey(repoURL, chartName, version)

	// Update existing entry
	if elem, ok := c.cache[key]; ok {
		c.lru.MoveToFront(elem)
		elem.Value.(*chartEntry).chart = chart
		return
	}

	// Add new entry
	entry := &chartEntry{key: key, chart: chart}
	elem := c.lru.PushFront(entry)
	c.cache[key] = elem

	// Evict oldest if over capacity
	for c.lru.Len() > c.capacity {
		oldest := c.lru.Back()
		if oldest != nil {
			c.lru.Remove(oldest)
			delete(c.cache, oldest.Value.(*chartEntry).key)
		}
	}
}

// Clear removes all entries from the cache.
func (c *ChartCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*list.Element)
	c.lru.Init()
}

// Size returns the current number of cached charts.
func (c *ChartCache) Size() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.lru.Len()
}

func makeChartKey(repoURL, chartName, version string) string {
	return repoURL + "\x00" + chartName + "\x00" + version
}

// repoLockManager manages per-repository locks for concurrent access.
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
func (m *repoLockManager) lock(repoURL string) func() {
	m.mu.Lock()
	rl, ok := m.locks[repoURL]
	if !ok {
		rl = &repoLock{}
		m.locks[repoURL] = rl
	}
	rl.refCount++
	m.mu.Unlock()

	rl.mu.Lock()

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
