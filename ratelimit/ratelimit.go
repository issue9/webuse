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
)

// GenFunc 用于生成用户唯一 ID 的函数，用于区分令牌桶所属的用户
type GenFunc func(*http.Request) (string, error)

// Store 存储 Bucket 的接口
type Store interface {
	// 设置或是添加指定名称的令牌桶
	Set(name string, val *Bucket) error

	// 删除指定的令牌桶
	Delete(name string) error

	// 获取指定的令牌桶，为空则返回 nil
	Get(name string) *Bucket
}

// Ratelimit 提供操作 Bucket 的一系列服务
type Ratelimit struct {
	store    Store
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
func New(store Store, capacity int64, rate time.Duration, fn GenFunc, errlog *log.Logger) *Ratelimit {
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

	b := rate.store.Get(name)
	if b == nil {
		b = newBucket(rate.capacity, rate.rate)
		if err := rate.store.Set(name, b); err != nil {
			return nil, err
		}
	}

	return b, nil
}

// Transfer 将 oldName 的数据传送给 newName
func (rate *Ratelimit) Transfer(oldName, newName string) error {
	b := rate.store.Get(oldName)
	if b != nil {
		if err := rate.store.Delete(oldName); err != nil {
			return err
		}
	}

	return rate.store.Set(newName, b)
}

func (rate *Ratelimit) printError(err error) {
	if rate.errlog != nil {
		rate.errlog.Println(err)
	}
}

// Middleware 将当前中间件应用于 next
func (rate *Ratelimit) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := rate.bucket(r)

		if err != nil {
			rate.printError(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		allow := b.allow(1) // 先拿走令牌
		b.setHeader(w)      // 再设置报头

		if !allow {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// MiddlewareFunc 将当前中间件应用于 next
func (rate *Ratelimit) MiddlewareFunc(next func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return rate.Middleware(http.HandlerFunc(next))
}
