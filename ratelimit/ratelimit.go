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
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/issue9/cache"
	"github.com/issue9/web"
)

// GenFunc 用于生成用户唯一 ID 的函数，用于区分令牌桶所属的用户
type GenFunc func(*http.Request) (string, error)

// Ratelimit 提供操作 Bucket 的一系列服务
type Ratelimit struct {
	store    cache.Cache
	capacity int64
	rate     time.Duration
	genFunc  GenFunc
	errlog   *log.Logger
}

// GenIP 用于生成区分令牌桶的 IP 地址
func GenIP(r *http.Request) (string, error) {
	if len(r.RemoteAddr) == 0 {
		return "", errors.New("无法获取请求端的 IP 地址")
	}

	if r.RemoteAddr[0] == '[' { // IPv6 带端口
		if index := strings.Index(r.RemoteAddr, "]:"); index > 0 {
			return r.RemoteAddr[:index+1], nil
		}
		return r.RemoteAddr, nil
	}

	if index := strings.IndexByte(r.RemoteAddr, ':'); index > 0 {
		return r.RemoteAddr[:index], nil
	}
	return r.RemoteAddr, nil
}

// New 声明一个新的 Ratelimit
//
// rate 拿令牌的频率
// fn 为令牌桶名称的产生方法，默认为用户的访问 IP。
func New(store cache.Cache, capacity int64, rate time.Duration, fn GenFunc, errlog *log.Logger) *Ratelimit {
	if fn == nil {
		fn = GenIP
	}

	return &Ratelimit{
		store:    store,
		capacity: capacity,
		rate:     rate,
		genFunc:  fn,
		errlog:   errlog,
	}
}

// 获取与当前请求相对应的令牌桶
func (rate *Ratelimit) bucket(r *http.Request) (*Bucket, error) {
	name, err := rate.genFunc(r)
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

func (rate *Ratelimit) printError(err error) {
	if rate.errlog != nil {
		rate.errlog.Println(err)
	}
}

// Middleware 将当前中间件应用于 next
func (rate *Ratelimit) Middleware(next web.HandlerFunc) web.HandlerFunc {
	return func(ctx *web.Context) *web.Response {
		b, err := rate.bucket(ctx.Request())
		if err != nil {
			rate.printError(err)
			return web.Status(http.StatusInternalServerError)
		}

		if b.allow(1) {
			return b.setHeader(next(ctx))
		}
		return b.setHeader(web.Status(http.StatusTooManyRequests))
	}
}
