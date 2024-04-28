// SPDX-FileCopyrightText: 2015-2024 caixw
//
// SPDX-License-Identifier: MIT

package jwt

import (
	"time"

	"github.com/issue9/web"
)

type (
	// Blocker 判断令牌是否被丢弃
	//
	// 在某些情况下，需要强制用户的令牌不再可用，可以使用此接口。
	Blocker[T Claims] interface {
		// BlockToken 拉黑令牌
		//
		// token 令牌；
		// refresh 是否为刷新令牌；
		BlockToken(token string, refresh bool) error

		// TokenIsBlocked 令牌是否已被丢弃
		TokenIsBlocked(string) bool

		// ClaimsIsBlocked 根据 Claims 判断是否已经丢弃
		//
		// 这是对令牌解码之后的阻断行为，性能上有解码的开销，便是相对来说也更加的灵活，
		// 比如要禁用某一用户所有签发的令牌，或是为某一设备签发的令牌等，
		// 只要 T 类型中带的字段均可作为判断依据。
		ClaimsIsBlocked(T) bool
	}

	cacheBlocker[T Claims] struct {
		accessTTL  time.Duration
		refreshTTL time.Duration
		cache      web.Cache
	}
)

// NewCacheBlocker 声明基于 [web.Cache] 的 [Blocker] 实现
//
// access 和 refresh 表示拉黑的令牌在多少时间之后会被释放；
func NewCacheBlocker[T Claims](c web.Cache, access, refresh time.Duration) Blocker[T] {
	return &cacheBlocker[T]{
		accessTTL:  access,
		refreshTTL: refresh,
		cache:      c,
	}
}

func (d *cacheBlocker[T]) BlockToken(token string, refresh bool) error {
	ttl := d.accessTTL
	if refresh {
		ttl = d.refreshTTL
	}
	return d.cache.Set(token, 1, ttl)
}

func (d *cacheBlocker[T]) TokenIsBlocked(token string) bool { return d.cache.Exists(token) }

func (d *cacheBlocker[T]) ClaimsIsBlocked(T) bool { return false }
