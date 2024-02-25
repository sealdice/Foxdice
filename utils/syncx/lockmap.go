package syncx

import "sync"

func NewRWMap[K comparable, V any]() *RWMap[K, V] {
	return &RWMap[K, V]{
		l: sync.RWMutex{},
		m: map[K]V{},
	}
}

type RWMap[K comparable, V any] struct {
	l sync.RWMutex
	m map[K]V
}

func (r *RWMap[K, V]) Load(key K) (V, bool) {
	r.l.RLock()
	v, ok := r.m[key]
	r.l.RUnlock()
	return v, ok
}

func (r *RWMap[K, V]) Store(key K, val V) {
	r.l.Lock()
	r.m[key] = val
	r.l.Unlock()
}
func (r *RWMap[K, V]) Delete(key K) {
	r.l.Lock()
	delete(r.m, key)
	r.l.Unlock()
}
