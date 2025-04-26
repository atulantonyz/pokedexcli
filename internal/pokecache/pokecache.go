package pokecache

import (
	"fmt"
	"sync"
	"time"
)

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

type Cache struct {
	interval     time.Duration
	CacheEntries map[string]cacheEntry
	mu           *sync.Mutex
}

func NewCache(interval time.Duration) Cache {
	c := Cache{
		interval:     interval,
		CacheEntries: make(map[string]cacheEntry),
		mu:           &sync.Mutex{},
	}
	go c.reapLoop()
	return c
}

func (c Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	new_entry := cacheEntry{}
	new_entry.createdAt = time.Now()
	new_entry.val = val
	fmt.Println("url : " + key + " Added to Cache")
	c.CacheEntries[key] = new_entry
}

func (c Cache) Get(key string) (val []byte, found bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	cache_entry, ok := c.CacheEntries[key]
	if !ok {
		return nil, false
	}
	fmt.Println("url : " + key + " Retrieved from Cache")
	return cache_entry.val, true
}

func (c Cache) reapLoop() {
	ticker := time.NewTicker(c.interval)
	for {
		t := <-ticker.C
		c.mu.Lock()
		for k := range c.CacheEntries {
			if c.CacheEntries[k].createdAt.Add(c.interval).Before(t) {
				delete(c.CacheEntries, k)
			}
		}
		c.mu.Unlock()
	}
}
