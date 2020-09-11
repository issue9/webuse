// SPDX-License-Identifier: MIT

package ratelimit

import "sync"

type memory struct {
	lock    sync.RWMutex
	buckets map[string]*Bucket
}

// NewMemory 声明一个内存类型的 Store
func NewMemory(capacity int64) Store {
	return &memory{
		buckets: make(map[string]*Bucket, capacity),
	}
}

func (mem *memory) Set(name string, b *Bucket) error {
	mem.lock.Lock()
	mem.buckets[name] = b
	mem.lock.Unlock()

	return nil
}

func (mem *memory) Delete(name string) error {
	mem.lock.Lock()
	delete(mem.buckets, name)
	mem.lock.Unlock()

	return nil
}

func (mem *memory) Get(name string) *Bucket {
	mem.lock.RLock()
	b := mem.buckets[name]
	mem.lock.RUnlock()

	return b
}
