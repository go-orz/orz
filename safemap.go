package orz

import "sync"

func NewSafeMap[K comparable, T any]() *SafeMap[K, T] {
	m := &SafeMap[K, T]{
		m: make(map[K]T),
	}
	return m
}

type SafeMap[K comparable, T any] struct {
	mutex sync.Mutex
	m     map[K]T
}

func (v *SafeMap[K, T]) Get(key K) (t T, ok bool) {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	t, ok = v.m[key]
	return
}

func (v *SafeMap[K, T]) Set(key K, av T) {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	v.m[key] = av
}

func (v *SafeMap[K, T]) Delete(key K) {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	delete(v.m, key)
}
