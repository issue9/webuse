// SPDX-License-Identifier: MIT

// Package ratelimit 提供了 X-Rate-Limit 功能的中间件
//
//  X-Rate-Limit-Limit: 同一个时间段所允许的请求的最大数目;
//  X-Rate-Limit-Remaining: 在当前时间段内剩余的请求的数量;
//  X-Rate-Limit-Reset: 为了得到最大请求数所等待的秒数。
//
//  store := NewMemory(...)
//  srv := New(store)
//  h = srv.Middleware(h)
package ratelimit

import (
	"errors"
	"net/http"
	"time"

	"github.com/issue9/cache"
	"github.com/issue9/web"
	"github.com/issue9/web/server"
)

// GenFunc 用于生成用户唯一 ID 的函数
//
// 用于区分令牌桶所属的用户
type GenFunc func(*web.Context) (string, error)

// Ratelimit 提供操作 Bucket 的一系列服务
type Ratelimit struct {
	store    cache.Access
	capacity int64
	rate     time.Duration
	genFunc  GenFunc
	errlog   web.Logger
}

// GenIP 用于生成区分令牌桶的 IP 地址
func GenIP(ctx *web.Context) (string, error) {
	if ip := ctx.ClientIP(); ip != "" {
		return ip, nil
	}
	return "", errors.New("ratelimit: 无法为请求生成唯一标记！")
}

// New 声明一个新的 Ratelimit
//
// rate 拿令牌的频率；
// fn 为令牌桶名称的产生方法，默认为用户的 IP；
func New(s *web.Server, prefix string, capacity int64, rate time.Duration, fn GenFunc) *Ratelimit {
	if fn == nil {
		fn = GenIP
	}

	return &Ratelimit{
		store:    cache.Prefix(prefix, s.Cache()),
		capacity: capacity,
		rate:     rate,
		genFunc:  fn,
		errlog:   s.Logs().ERROR(),
	}
}

// 获取与当前请求相对应的令牌桶
func (rate *Ratelimit) bucket(ctx *web.Context) (*Bucket, error) {
	name, err := rate.genFunc(ctx)
	if err != nil {
		return nil, err
	}

	b, err := rate.store.Get(name)
	if err == cache.ErrCacheMiss {
		b = newBucket(rate.capacity, rate.rate)
		if err := rate.store.Set(name, b, cache.Forever); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return b.(*Bucket), nil
}

// Transfer 将 oldName 的数据传送给 newName
func (rate *Ratelimit) Transfer(oldName, newName string) error {
	b, err := rate.store.Get(oldName)
	if err != nil && err != cache.ErrCacheMiss {
		return err
	}

	if b != nil {
		if err := rate.store.Delete(oldName); err != nil {
			return err
		}
		return rate.store.Set(newName, b, cache.Forever)
	}

	return nil
}

// Middleware 将当前中间件应用于 next
func (rate *Ratelimit) Middleware(next web.HandlerFunc) web.HandlerFunc {
	return func(ctx *web.Context) web.Responser {
		b, err := rate.bucket(ctx)
		if err != nil {
			rate.errlog.Error(err)
			return server.Status(http.StatusInternalServerError)
		}

		if b.allow(1) {
			return next(b.setHeader(ctx))
		}
		b.setHeader(ctx)
		return server.Status(http.StatusTooManyRequests)
	}
}
