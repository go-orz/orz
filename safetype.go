package orz

import "sync"

type SafeType[T any] struct {
	val   T
	mutex sync.RWMutex
}

func (v *SafeType[T]) Get() T {
	v.mutex.RLock()
	defer v.mutex.RUnlock()
	return v.val
}

func (v *SafeType[T]) Set(val T) {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	v.val = val
}
