// SPDX-FileCopyrightText: 2015-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package ratelimit API 限流中间件
//
// 这是以用户或是客户端为单位的限流中间件，并不能按 API 进行细化的限流。
//
// NOTE: 所有数据保存在 [web.Cache] 之中，缓存服务重启后数据也将重置。
package ratelimit

import (
	"errors"
	"strconv"
	"time"

	"github.com/issue9/cache"
	"github.com/issue9/mux/v8/header"
	"github.com/issue9/web"
)

// GenFunc 用于生成用户唯一 ID 的函数
//
// 用于区分令牌桶所属的用户
type GenFunc = func(*web.Context) (string, error)

type ratelimit struct {
	store       web.Cache
	capacity    uint64
	rate        time.Duration
	rateSeconds int
	gen         GenFunc

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
// rate 拿令牌的频率；
// gen 为令牌桶名称的产生方法，默认为用户的 IP；
// headers 自定义报头名称，可以指定以下值：
//   - X-Rate-Limit-Limit: 同一个时间段所允许的请求的最大数目;
//   - X-Rate-Limit-Remaining: 在当前时间段内剩余的请求的数量;
//   - X-Rate-Limit-Reset: 为了得到最大请求数所需等待的 UNIX 时间。
func New(c web.Cache, capacity uint64, rate time.Duration, gen GenFunc, headers map[string]string) web.Middleware {
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

	return &ratelimit{
		store:       c,
		capacity:    capacity,
		rate:        rate,
		rateSeconds: int(rate.Seconds()),
		gen:         gen,

		limit:     headers[header.XRateLimitLimit],
		remaining: headers[header.XRateLimitRemaining],
		reset:     headers[header.XRateLimitReset],
	}
}

func (rate *ratelimit) Middleware(next web.HandlerFunc) web.HandlerFunc {
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

// 是否允许当前请求
//
// 如果允许，则返回当前可用的数量。
func (rate *ratelimit) allow(ctx *web.Context) (uint64, error) {
	name, err := rate.gen(ctx)
	if err != nil {
		return 0, err
	}
	lastName := name + "_last"
	_, setter, found, err := rate.store.Counter(name+"_cnt", rate.rate)
	if err != nil {
		return 0, err
	}
	if !found {
		setter(int(rate.capacity))
	}

	size, err := setter(-1) // 先扣点，保证多并发情况下不会有问题。
	if err != nil {
		return 0, err
	}

	now := ctx.Begin()
	var last time.Time
	switch err = rate.store.Get(lastName, &last); {
	case errors.Is(err, cache.ErrCacheMiss()):
		last = now
	case err != nil:
		ctx.Logs().ERROR().Error(err)
		return size, nil // 无法确定最后日期，就以当前的数量为准
	}

	dur := now.Sub(last)               // 从上次拿令牌到现在的时间段
	cnt := uint64(dur / rate.rate)     // 计算这段时间内需要增加的令牌
	cnt = min(cnt, rate.capacity-size) // 不超过增加的量
	if cnt > 0 {
		if _, err = setter(int(cnt)); err != nil {
			ctx.Logs().ERROR().Error(err)
			return size + cnt, nil
		}
	}

	if err = rate.store.Set(lastName, now, cache.Forever); err != nil {
		ctx.Logs().ERROR().Error(err)
		return size + cnt, nil
	}

	return size + cnt, nil
}

func (rate *ratelimit) setHeader(ctx *web.Context, size uint64) {
	t := (rate.capacity - size) * uint64(rate.rateSeconds)
	rest := ctx.Begin().Unix() + int64(t)

	h := ctx.Header()
	h.Set(rate.limit, strconv.FormatUint(rate.capacity, 10))
	h.Set(rate.remaining, strconv.FormatUint(size, 10))
	h.Set(rate.reset, strconv.FormatInt(rest, 10))
}
