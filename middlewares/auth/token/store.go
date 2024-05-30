// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package token

import (
	"errors"
	"time"

	"github.com/issue9/cache"
	"github.com/issue9/web"
)

// UserData 作为用户数据存储时需要实现的接口
type UserData interface {
	// GetUID 获取当前数据的关联用户 ID
	GetUID() string
}

// Item 令牌关联的数据
type Item[V UserData] struct {
	Access   string // 如果是刷新令牌，此值关联着访问令牌，否则为空。
	UserData V      // 令牌关联的数据
}

// Store 令牌的存储接口
type Store[V UserData] interface {
	// 保存数据
	//
	// token 令牌的值；
	// data 与令牌关联的数据；
	// ttl 过期时间；
	Save(token string, data Item[V], ttl time.Duration) error

	// 通过令牌删除关联数据
	DeleteToken(token string) error

	// 通过 [UserData.GetUID] 删除关联数据
	DeleteUID(uid string) error

	// 获取与令牌 token 关联的数据
	//
	// 如果不存在，应该返回 nil
	Get(token string) (Item[V], bool, error)
}

type cacheStore[T UserData] struct {
	tokenItem  web.Cache // token: item
	uidAccess  web.Cache // uid: access token
	uidRefresh web.Cache // uid: refresh token
}

// NewCacheStore 声明基于 [web.Cache] 的 [Store] 实现
func NewCacheStore[T UserData](c web.Cache) Store[T] {
	return &cacheStore[T]{
		tokenItem:  cache.Prefix(c, "i_"),
		uidAccess:  cache.Prefix(c, "a_"),
		uidRefresh: cache.Prefix(c, "r_"),
	}
}

func (s *cacheStore[T]) Save(token string, v Item[T], ttl time.Duration) error {
	c := s.uidAccess
	if v.Access != "" {
		c = s.uidRefresh
	}
	return errors.Join(s.tokenItem.Set(token, v, ttl), c.Set(v.UserData.GetUID(), token, ttl))
}

func (s *cacheStore[T]) DeleteToken(token string) error {
	item, err := cache.Get[Item[T]](s.tokenItem, token)
	if err != nil {
		return err
	}

	c := s.uidAccess
	if item.Access != "" {
		c = s.uidRefresh
	}
	return errors.Join(s.tokenItem.Delete(token), c.Delete(item.UserData.GetUID()))
}

func (s *cacheStore[T]) DeleteUID(uid string) error {
	access, err := cache.Get[string](s.uidAccess, uid)
	if err != nil {
		return err
	}
	refresh, err := cache.Get[string](s.uidRefresh, uid)
	if err != nil {
		return err
	}
	return errors.Join(s.uidAccess.Delete(uid), s.uidRefresh.Delete(uid), s.tokenItem.Delete(access), s.tokenItem.Delete(refresh))
}

func (s *cacheStore[T]) Get(token string) (Item[T], bool, error) {
	v, err := cache.Get[Item[T]](s.tokenItem, token)
	switch {
	case errors.Is(err, cache.ErrCacheMiss()):
		return Item[T]{}, false, nil
	case err != nil:
		return Item[T]{}, false, err
	}
	return v, true, nil
}
