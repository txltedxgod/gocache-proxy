package cache

import (
	"net/http"
	"testing"
	"time"
)

func TestLRUCache_BasicSetGet(t *testing.T) {
	c := NewLRUCache(2)

	item1 := &ResponseItem{
		StatusCode: http.StatusOK,
		Body:       []byte("response 1"),
		ExpiresAt:  time.Now().Add(1 * time.Minute),
	}

	c.Set("key1", item1)

	got, found := c.Get("key1")
	if !found {
		t.Fatalf("Expected key1 to be found")
	}
	if string(got.Body) != "response 1" {
		t.Errorf("Expected 'response 1', got %s", string(got.Body))
	}
}

func TestLRUCache_Eviction(t *testing.T) {
	c := NewLRUCache(2)

	c.Set("k1", &ResponseItem{Body: []byte("1"), ExpiresAt: time.Now().Add(time.Minute)})
	c.Set("k2", &ResponseItem{Body: []byte("2"), ExpiresAt: time.Now().Add(time.Minute)})

	// Access k1 to make k2 least recently used
	c.Get("k1")

	// Insert k3, should evict k2
	c.Set("k3", &ResponseItem{Body: []byte("3"), ExpiresAt: time.Now().Add(time.Minute)})

	if _, found := c.Get("k2"); found {
		t.Errorf("Expected k2 to be evicted")
	}

	if _, found := c.Get("k1"); !found {
		t.Errorf("Expected k1 to still exist")
	}
}

func TestLRUCache_Expiration(t *testing.T) {
	c := NewLRUCache(5)

	c.Set("expiring", &ResponseItem{
		Body:      []byte("short lived"),
		ExpiresAt: time.Now().Add(50 * time.Millisecond),
	})

	time.Sleep(70 * time.Millisecond)

	if _, found := c.Get("expiring"); found {
		t.Errorf("Expected expiring key to be expired and not found")
	}
}
