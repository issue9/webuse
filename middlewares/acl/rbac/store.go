// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package rbac

import (
	"errors"
	"fmt"
	"slices"

	"github.com/issue9/cache"
	"github.com/issue9/web"
)

// Store 存储接口
type Store[T comparable] interface {
	// Load 加载所有的角色
	Load() (map[string]*Role[T], error)

	// Del 删除指定角色
	//
	// 如果角色下面还有子角色或是用户时，不应该删除。
	Del(string) error

	// Set 修改角色信息
	//
	// 如果修改了角色可访问的资源列表，应该检测子角色是否拥有该资源。
	Set(*Role[T]) error

	// Add 添加角色信息
	Add(*Role[T]) error

	// Exists 是否存在指定 id 的角色
	Exists(string) (bool, error)
}

type cacheStore[T comparable] struct {
	s web.Server
	c web.Cache
}

// NewCacheStore 声明基于 [web.Cache] 的 [Store] 实现
//
// NOTE: 缓存是易失性的，不太具备实用性，可用于测试。
func NewCacheStore[T comparable](s web.Server, prefix string) Store[T] {
	c := web.NewCache(prefix, s.Cache())
	c.Set("", []string{}, cache.Forever) // 防止 cache miss

	return &cacheStore[T]{
		s: s,
		c: c,
	}
}

func (s *cacheStore[T]) getKeys() ([]string, error) {
	keys := []string{}
	if err := s.c.Get("", &keys); err != nil {
		return nil, err
	}
	return keys, nil
}

func (s *cacheStore[T]) setKeys(keys []string) error { return s.c.Set("", keys, cache.Forever) }

func (s *cacheStore[T]) Load() (map[string]*Role[T], error) {
	keys, err := s.getKeys()
	if err != nil {
		return nil, err
	}

	roles := make([]*Role[T], 0, len(keys))
	for _, item := range keys {
		r := &Role[T]{}
		if err := s.c.Get(item, r); err != nil {
			return nil, err
		}
		roles = append(roles, r)
	}

	ms := make(map[string]*Role[T], len(roles))
	for _, role := range roles {
		if role.Parent != "" {
			if index := slices.IndexFunc(roles, func(r *Role[T]) bool { return r.ID == role.Parent }); index >= 0 {
				role.parent = roles[index]
			}
		}
		ms[role.ID] = role
	}

	return ms, nil
}

func (s *cacheStore[T]) Del(id string) error {
	keys, err := s.getKeys()
	if err != nil {
		return err
	}

	if err := s.c.Delete(id); err != nil && !errors.Is(err, cache.ErrCacheMiss()) {
		return err
	}

	if index := slices.Index(keys, id); index >= 0 {
		keys = slices.Delete(keys, index, index+1)
		return s.setKeys(keys)
	}
	return nil
}

func (s *cacheStore[T]) Set(role *Role[T]) error {
	return s.c.Set(role.ID, role, cache.Forever)
}

func (s *cacheStore[T]) Add(role *Role[T]) error {
	keys, err := s.getKeys()
	if err != nil {
		return err
	}

	if slices.Index(keys, role.ID) >= 0 {
		// 由 server.UniqueID() 保证 role.ID 唯一性，如果不唯一，肯定是代码级别的错误。
		panic(fmt.Sprintf("角色 %s 已经存在", role.ID))
	}

	if err := s.c.Set(role.ID, role, cache.Forever); err != nil {
		return err
	}

	return s.setKeys(append(keys, role.ID))
}

func (s *cacheStore[T]) Exists(id string) (bool, error) {
	keys, err := s.getKeys()
	if err != nil && !errors.Is(err, cache.ErrCacheMiss()) {
		return false, err
	}
	return slices.Index(keys, id) >= 0, nil
}
