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
	// Load 加载 gid 下的所有角色
	Load(gid string) (map[string]*Role[T], error)

	// Del 删除指定角色
	//
	// 如果角色下面还有子角色或是用户时，不应该删除。
	Del(gid, roleID string) error

	// Set 修改角色信息
	//
	// 如果修改了角色可访问的资源列表，应该检测子角色是否拥有该资源。
	Set(gid string, r *Role[T]) error

	// Add 添加角色信息
	//
	// r.id 是由 [Server.UniqueID] 保证不重复的，如果出现重复的情况，应该直接 panic.
	Add(gid string, r *Role[T]) error
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

func (s *cacheStore[T]) getRoleIDs(gid string) ([]string, error) {
	ids := []string{}
	err := s.c.Get(gid, &ids)
	if errors.Is(err, cache.ErrCacheMiss()) {
		return ids, nil
	} else if err != nil {
		return nil, err
	}
	return ids, nil
}

func (s *cacheStore[T]) setRoleIDs(gid string, ids []string) error {
	return s.c.Set(gid, ids, cache.Forever)
}

func (s *cacheStore[T]) Load(gid string) (map[string]*Role[T], error) {
	ids, err := s.getRoleIDs(gid)
	if err != nil {
		return nil, err
	}

	roles := make([]*Role[T], 0, len(ids))
	for _, item := range ids {
		r := &Role[T]{}
		if err := s.c.Get(gid+item, r); err != nil {
			return nil, err
		}
		roles = append(roles, r)
	}

	// 确定 parent 关系
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

func (s *cacheStore[T]) Del(gid, id string) error {
	ids, err := s.getRoleIDs(gid)
	if err != nil {
		return err
	}

	if err := s.c.Delete(gid + id); err != nil && !errors.Is(err, cache.ErrCacheMiss()) {
		return err
	}

	if index := slices.Index(ids, id); index >= 0 {
		ids = slices.Delete(ids, index, index+1)
		return s.setRoleIDs(gid, ids)
	}
	return nil
}

func (s *cacheStore[T]) Set(gid string, role *Role[T]) error {
	return s.c.Set(gid+role.ID, role, cache.Forever)
}

func (s *cacheStore[T]) Add(gid string, role *Role[T]) error {
	ids, err := s.getRoleIDs(gid)
	if err != nil {
		return err
	}

	if slices.Index(ids, role.ID) >= 0 {
		// 由 server.UniqueID() 保证 role.ID 唯一性，如果不唯一，肯定是代码级别的错误。
		panic(fmt.Sprintf("角色 %s 已经存在", role.ID))
	}

	if err := s.c.Set(gid+role.ID, role, cache.Forever); err != nil {
		return err
	}

	return s.setRoleIDs(gid, append(ids, role.ID))
}
