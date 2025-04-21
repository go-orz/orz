package orz

import (
	"iter"
	"maps"
	"sync"
)

// SafeMap is a thread-safe map.
type SafeMap[K comparable, V any] struct {
	mu sync.RWMutex
	m  map[K]V
}

// NewSafeMap returns a new SafeMap.
func NewSafeMap[K comparable, V any]() *SafeMap[K, V] {
	return &SafeMap[K, V]{
		m: make(map[K]V),
	}
}

// Get returns the value associated with the given key.
func (m *SafeMap[K, V]) Get(key K) (V, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.m[key]
	return v, ok
}

// Set sets the value associated with the given key.
func (m *SafeMap[K, V]) Set(key K, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.m[key] = value
}

// Delete deletes the value associated with the given key.
func (m *SafeMap[K, V]) Delete(key K) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.m, key)
}

// Keys returns a sequence of all keys in the map.
func (m *SafeMap[K, V]) Keys() iter.Seq[K] {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return maps.Keys(m.m)
}

// Values returns a sequence of all values in the map.
func (m *SafeMap[K, V]) Values() iter.Seq[V] {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return maps.Values(m.m)
}

func (m *SafeMap[K, V]) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.m)
}
