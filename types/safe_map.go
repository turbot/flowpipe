package types

import "sync"

type SafeMap[K comparable, V any] struct {
	mu    sync.Mutex
	items map[K]V
}

// NewSafeMap creates a new instance of a SafeMap.
func NewSafeMap[K comparable, V any]() *SafeMap[K, V] {
	return &SafeMap[K, V]{
		items: make(map[K]V),
	}
}

// Set sets a key-value pair in the SafeMap.
func (sm *SafeMap[K, V]) Set(key K, value V) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.items[key] = value
}

// Get retrieves a value for a key from the SafeMap.
func (sm *SafeMap[K, V]) Get(key K) (V, bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	value, ok := sm.items[key]
	return value, ok
}

// Delete removes a key-value pair from the SafeMap.
func (sm *SafeMap[K, V]) Delete(key K) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.items, key)
}
