// SPDX-FileCopyrightText: 2015-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package ratelimit API 限流中间件
//
// 这是以用户或是客户端为单位的限流中间件，如果需要按路由进行限流，需为每个路由指定一个实例。
//
// NOTE: 所有数据保存在 [web.Cache] 之中，缓存服务重启后数据也将重置。
package ratelimit

import (
	"errors"
	"slices"
	"strconv"
	"time"

	"github.com/issue9/cache"
	"github.com/issue9/mux/v9/header"
	"github.com/issue9/web"
)

// GenFunc 用于生成用户唯一 ID 的函数
//
// 用于区分令牌桶所属的用户
type GenFunc = func(*web.Context) (string, error)

type Ratelimit struct {
	store    web.Cache
	capacity uint64
	rate     time.Duration
	ttl      time.Duration
	gen      GenFunc
	unlimit  []string

	limit, remaining, reset string
}

// GenIP 用于生成区分令牌桶的 IP 地址
func GenIP(ctx *web.Context) (string, error) {
	if ip := ctx.ClientIP(); ip != "" {
		return ip, nil
	}
	return "", web.NewLocaleError("can not get the ip")
}

// New 声明 API 限流的中间件
//
// capacity 桶的容量；
// rate 发放令牌的时间间隔；
// gen 为令牌桶名称的产生方法，默认为用户的 IP；
// headers 自定义报头名称，可以指定以下值：
//   - X-Rate-Limit-Limit: 同一个时间段所允许的请求的最大数目;
//   - X-Rate-Limit-Remaining: 在当前时间段内剩余的请求的数量;
//   - X-Rate-Limit-Reset: 为了得到最大请求数所需等待的 UNIX 时间。
func New(c web.Cache, capacity uint64, rate time.Duration, gen GenFunc, headers map[string]string) *Ratelimit {
	ttl := rate * time.Duration(capacity)
	if ttl < time.Second {
		panic("capacity*rate 必须大于 1 秒")
	}

	if gen == nil {
		gen = GenIP
	}

	if headers == nil {
		headers = map[string]string{}
	}

	if _, found := headers[header.XRateLimitLimit]; !found {
		headers[header.XRateLimitLimit] = header.XRateLimitLimit
	}
	if _, found := headers[header.XRateLimitRemaining]; !found {
		headers[header.XRateLimitRemaining] = header.XRateLimitRemaining
	}
	if _, found := headers[header.XRateLimitReset]; !found {
		headers[header.XRateLimitReset] = header.XRateLimitReset
	}

	return &Ratelimit{
		store:    c,
		capacity: capacity,
		rate:     rate,
		ttl:      ttl,
		gen:      gen,
		unlimit:  make([]string, 0, 10),

		limit:     headers[header.XRateLimitLimit],
		remaining: headers[header.XRateLimitRemaining],
		reset:     headers[header.XRateLimitReset],
	}
}

func buildID(method, path, router string) string { return router + method + path }

func (rate *Ratelimit) Middleware(next web.HandlerFunc, method, path, router string) web.HandlerFunc {
	if slices.Index(rate.unlimit, buildID(method, path, router)) >= 0 { //
		return next
	}

	return func(ctx *web.Context) web.Responser {
		size, err := rate.allow(ctx)
		if err != nil {
			return ctx.Problem(web.ProblemInternalServerError)
		}

		rate.setHeader(ctx, size)
		if size > 0 {
			return next(ctx)
		}
		return ctx.Problem(web.ProblemTooManyRequests)
	}
}

// Unlimit 返回一个脱离当前限制的中间件
func (rate *Ratelimit) Unlimit() web.Middleware {
	return web.MiddlewareFunc(func(next func(*web.Context) web.Responser, method, path, router string) func(*web.Context) web.Responser {
		rate.unlimit = append(rate.unlimit, buildID(method, path, router))
		return next
	})
}

// 是否允许当前请求
//
// 如果允许，则返回当前可用的数量。
func (rate *Ratelimit) allow(ctx *web.Context) (uint64, error) {
	name, err := rate.gen(ctx)
	if err != nil {
		return 0, err
	}
	_, setter, found, err := rate.store.Counter(name+"_cnt", rate.ttl)
	if err != nil {
		return 0, err
	}
	if !found {
		if _, err = setter(int(rate.capacity)); err != nil {
			return 0, err
		}
	}

	size, err := setter(-1) // 先扣点，保证多并发情况下不会有问题。
	if err != nil {
		return 0, err
	}

	// 根据最后一次请求时间，补发这段时间内需要的令牌数量。

	lastName := name + "_last"
	now := ctx.Begin()
	var last time.Time
	switch err = rate.store.Get(lastName, &last); {
	case errors.Is(err, cache.ErrCacheMiss()):
		last = now
	case err != nil:
		ctx.Logs().ERROR().Error(err)
		return size, nil // 无法确定最后日期，就以当前的数量为准
	}

	dur := now.Sub(last)                 // 从上次拿令牌到现在的时间段
	incr := uint64(dur / rate.rate)      // 计算这段时间内需要增加的令牌
	incr = min(incr, rate.capacity-size) // 不超过增加的量
	if incr > 0 {
		if _, err = setter(int(incr)); err != nil {
			ctx.Logs().ERROR().Error(err)
			return size + incr, nil
		}
	}

	if err = rate.store.Set(lastName, now, rate.ttl); err != nil {
		ctx.Logs().ERROR().Error(err)
		return size + incr, nil
	}

	return size + incr, nil
}

func (rate *Ratelimit) setHeader(ctx *web.Context, size uint64) {
	t := time.Duration(rate.capacity-size) * rate.rate
	rest := ctx.Begin().Add(t).Unix()

	h := ctx.Header()
	h.Set(rate.limit, strconv.FormatUint(rate.capacity, 10))
	h.Set(rate.remaining, strconv.FormatUint(size, 10))
	h.Set(rate.reset, strconv.FormatInt(rest, 10))
}
