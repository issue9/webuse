// SPDX-FileCopyrightText: 2015-2024 caixw
//
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
	"strconv"
	"time"

	"github.com/issue9/cache"
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
	genFunc     GenFunc
}

// GenIP 用于生成区分令牌桶的 IP 地址
func GenIP(ctx *web.Context) (string, error) {
	if ip := ctx.ClientIP(); ip != "" {
		return ip, nil
	}
	return "", web.NewLocaleError("can not get the ip")
}

// New 声明 [Ratelimit] 对象
//
// rate 拿令牌的频率；
// fn 为令牌桶名称的产生方法，默认为用户的 IP；
func New(c web.Cache, capacity uint64, rate time.Duration, fn GenFunc) web.Middleware {
	if fn == nil {
		fn = GenIP
	}

	return &ratelimit{
		store:       c,
		capacity:    capacity,
		rate:        rate,
		rateSeconds: int(rate.Seconds()),
		genFunc:     fn,
	}
}

// Middleware 将当前中间件应用于 next
func (rate *ratelimit) Middleware(next web.HandlerFunc) web.HandlerFunc {
	return func(ctx *web.Context) web.Responser {
		size, err := rate.allow(ctx)
		if err != nil {
			return ctx.Problem(web.ProblemInternalServerError)
		}

		if size > 0 {
			setHeader(rate, ctx, size)
			return next(ctx)
		}
		setHeader(rate, ctx, size)
		return ctx.Problem(web.ProblemTooManyRequests)
	}
}

// 是否允许当前请求
//
// 如果允许，则返回当前可用的数量。
func (rate *ratelimit) allow(ctx *web.Context) (uint64, error) {
	name, err := rate.genFunc(ctx)
	if err != nil {
		return 0, err
	}
	cntName := name + "_cnt"
	lastName := name + "_time"
	counter := rate.store.Counter(cntName, rate.capacity, rate.rate)

	size, err := counter.Decr(1) // 先扣点，保证多并发情况下不会有问题。
	if err != nil {
		return 0, err
	}

	now := ctx.Begin()
	var last time.Time
	err = rate.store.Get(lastName, &last)
	switch {
	case errors.Is(err, cache.ErrCacheMiss()):
		last = now
	case err != nil:
		ctx.Logs().ERROR().Error(err)
		return size, nil // 无法确定最后日期，就以当前的数量为准
	}

	dur := now.Sub(last)           // 从上次拿令牌到现在的时间段
	cnt := uint64(dur / rate.rate) // 计算这段时间内需要增加的令牌
	if max := rate.capacity - size; cnt > max {
		cnt = max
	}
	if cnt > 0 {
		if _, err = counter.Incr(cnt); err != nil {
			ctx.Logs().ERROR().Error(err)
			return size + cnt, nil
		}
	}

	err = rate.store.Set(lastName, now, cache.Forever)
	if err != nil && !errors.Is(err, cache.ErrCacheMiss()) {
		ctx.Logs().ERROR().Error(err)
		return size + cnt, nil
	}

	return size + cnt, nil
}

func setHeader(rate *ratelimit, ctx *web.Context, size uint64) {
	t := (rate.capacity - size) * uint64(rate.rateSeconds)
	rest := ctx.Begin().Unix() + int64(t)

	h := ctx.Header()
	h.Set("X-Rate-Limit-Limit", strconv.FormatUint(rate.capacity, 10))
	h.Set("X-Rate-Limit-Remaining", strconv.FormatUint(size, 10))
	h.Set("X-Rate-Limit-Reset", strconv.FormatInt(rest, 10))
}