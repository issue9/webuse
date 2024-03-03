// SPDX-FileCopyrightText: 2015-2024 caixw
//
// SPDX-License-Identifier: MIT

package session

import (
	"errors"
	"time"

	"github.com/issue9/cache"
	"github.com/issue9/web"
)

// Store session 的存储接口
type Store[T any] interface {
	// Delete 删除指定 id 的 session
	Delete(id string) error

	// Get 查找指定 id 的 session
	//
	// 如果不存在，则返回零值，并将该零值保存到当前对象。
	Get(id string) (*T, error)

	// Set 更新指定 id 的 session
	Set(id string, v *T) error
}

type cacheStore[T any] struct {
	ttl time.Duration
	c   web.Cache
}

// NewCacheStore 以 [web.Cache] 作为 session 的存储系统
func NewCacheStore[T any](c web.Cache, ttl time.Duration) Store[T] {
	return &cacheStore[T]{
		ttl: ttl,
		c:   c,
	}
}

func (s *cacheStore[T]) Delete(id string) error { return s.c.Delete(id) }

func (s *cacheStore[T]) Get(id string) (*T, error) {
	var v T
	err := s.c.Get(id, &v)
	if errors.Is(err, cache.ErrCacheMiss()) {
		if err := s.c.Set(id, &v, s.ttl); err != nil {
			return &v, err
		}
		return &v, nil
	}
	return &v, err
}

func (s *cacheStore[T]) Set(id string, v *T) error {
	return s.c.Set(id, v, s.ttl)
}
