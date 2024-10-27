package cache

import (
	"github.com/Scalingo/sclng-backend-test-v1/structs"
	"sync"
	"time"
)

type Cache struct {
	mu         sync.Mutex
	data       []*structs.GithubRepo
	timestamp  time.Time
	timeToLive time.Duration
}

func NewCache(timeToLive time.Duration) *Cache {
	return &Cache{
		data:       make([]*structs.GithubRepo, 0),
		timestamp:  time.Now(),
		timeToLive: timeToLive,
	}
}

func (cache *Cache) Set(data []*structs.GithubRepo) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.data = data
	cache.timestamp = time.Now()
}

func (cache *Cache) GetCacheData() ([]*structs.GithubRepo, bool) {
	// if cache data is empty or expired, return nothing
	if len(cache.data) != 0 && time.Since(cache.timestamp) < cache.timeToLive {
		return cache.data, true
	}
	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.data = make([]*structs.GithubRepo, 0)
	return nil, false
}
