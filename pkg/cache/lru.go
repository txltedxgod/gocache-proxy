package cache

import (
	"container/list"
	"net/http"
	"sync"
	"time"
)

// ResponseItem represents a cached HTTP response.
type ResponseItem struct {
	StatusCode int
	Header     http.Header
	Body       []byte
	ExpiresAt  time.Time
}

func (item *ResponseItem) IsExpired() bool {
	return time.Now().After(item.ExpiresAt)
}

type entry struct {
	key   string
	value *ResponseItem
}

// LRUCache is a concurrent, thread-safe LRU cache with expiration.
type LRUCache struct {
	mu       sync.RWMutex
	capacity int
	items    map[string]*list.Element
	evictList *list.List
}

func NewLRUCache(capacity int) *LRUCache {
	if capacity <= 0 {
		capacity = 1000
	}
	return &LRUCache{
		capacity:  capacity,
		items:     make(map[string]*list.Element),
		evictList: list.New(),
	}
}

// Get retrieves a response item from the cache.
func (c *LRUCache) Get(key string) (*ResponseItem, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, exists := c.items[key]
	if !exists {
		return nil, false
	}

	item := elem.Value.(*entry).value
	if item.IsExpired() {
		c.removeElement(elem)
		return nil, false
	}

	c.evictList.MoveToFront(elem)
	return item, true
}

// Set adds or updates a response item in the cache.
func (c *LRUCache) Set(key string, value *ResponseItem) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.items[key]; exists {
		c.evictList.MoveToFront(elem)
		elem.Value.(*entry).value = value
		return
	}

	if c.evictList.Len() >= c.capacity {
		c.evictOldest()
	}

	elem := c.evictList.PushFront(&entry{key: key, value: value})
	c.items[key] = elem
}

// Delete removes a specific key.
func (c *LRUCache) Delete(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.items[key]; exists {
		c.removeElement(elem)
		return true
	}
	return false
}

// Len returns the current number of cached items.
func (c *LRUCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.evictList.Len()
}

func (c *LRUCache) evictOldest() {
	elem := c.evictList.Back()
	if elem != nil {
		c.removeElement(elem)
	}
}

func (c *LRUCache) removeElement(elem *list.Element) {
	c.evictList.Remove(elem)
	kv := elem.Value.(*entry)
	delete(c.items, kv.key)
}
