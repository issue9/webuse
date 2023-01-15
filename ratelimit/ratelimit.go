// SPDX-License-Identifier: MIT

// Package ratelimit API 限流中间件
//
// 这是以用户或是客户端为单位的限流中间件，并不能按 API 进行细化的限流。
//
// 提供了对以下报头的支持：
// - X-Rate-Limit-Limit: 同一个时间段所允许的请求的最大数目;
// - X-Rate-Limit-Remaining: 在当前时间段内剩余的请求的数量;
// - X-Rate-Limit-Reset: 为了得到最大请求数所需等待的 UNIX 时间。
//
// 所有数据保存在 [web.Cache] 之中，缓存服务重启后数据也将重置。
package ratelimit

import (
	"errors"
	"time"

	"github.com/issue9/web"
	"github.com/issue9/web/cache"
)

// GenFunc 用于生成用户唯一 ID 的函数
//
// 用于区分令牌桶所属的用户
type GenFunc func(*web.Context) (string, error)

// Ratelimit 提供操作 Bucket 的一系列服务
type Ratelimit struct {
	store    web.Cache
	capacity int64
	rate     time.Duration
	genFunc  GenFunc
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
		store:    cache.Prefix(s.Cache(), prefix+"_"),
		capacity: capacity,
		rate:     rate,
		genFunc:  fn,
	}
}

// 获取与当前请求相对应的令牌桶
func (rate *Ratelimit) bucket(name string) (*Bucket, error) {
	b := &Bucket{}
	if err := rate.store.Get(name, b); errors.Is(err, cache.ErrCacheMiss()) {
		b = newBucket(rate.capacity, rate.rate)
		if err := rate.store.Set(name, b, cache.Forever); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return b, nil
}

// Transfer 将 oldName 的数据传送给 newName
func (rate *Ratelimit) Transfer(oldName, newName string) error {
	b := &Bucket{}
	err := rate.store.Get(oldName, b)
	switch {
	case errors.Is(err, cache.ErrCacheMiss()): // 不需要特殊处理
	case err != nil:
		return err
	default:
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
		name, err := rate.genFunc(ctx)
		if err != nil {
			return ctx.InternalServerError(err)
		}

		b, err := rate.bucket(name)
		if err != nil {
			return ctx.InternalServerError(err)
		}

		if b.allow(1) {
			rate.store.Set(name, b, cache.Forever)
			b.setHeader(ctx)
			return next(ctx)
		}
		b.setHeader(ctx)
		return ctx.Problem(web.ProblemTooManyRequests)
	}
}
