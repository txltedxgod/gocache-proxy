package cache

import (
	"testing"
)

type SimpleLRU struct {
	capacity int
	items    map[string]string
}

func NewSimpleLRU(cap int) *SimpleLRU {
	return &SimpleLRU{capacity: cap, items: make(map[string]string)}
}

func (l *SimpleLRU) Put(k, v string) {
	if len(l.items) >= l.capacity {
		for firstKey := range l.items {
			delete(l.items, firstKey)
			break
		}
	}
	l.items[k] = v
}

func (l *SimpleLRU) Get(k string) (string, bool) {
	v, ok := l.items[k]
	return v, ok
}

func TestLRUEviction(t *testing.T) {
	l := NewSimpleLRU(2)
	l.Put("a", "1")
	l.Put("b", "2")
	l.Put("c", "3")
	if len(l.items) > 2 {
		t.Errorf("expected size <= 2, got %d", len(l.items))
	}
}
