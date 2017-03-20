// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package ratelimit

import (
	"net/http"
	"strconv"
	"time"
)

type Bucket struct {
	capacity int64         // 上限
	rate     time.Duration // 每隔 rate 添加一个令牌
	tokens   int64         // 可用令牌数量
	last     time.Time     // 最后次添加令牌的时间
}

func newBucket(capacity int64, rate time.Duration) *Bucket {
	return &Bucket{
		capacity: capacity,
		rate:     rate,
		tokens:   capacity, // 默认为满格
		last:     time.Now(),
	}
}

func (b *Bucket) allow(n int64) bool {
	now := time.Now()

	if n > b.capacity {
		b.tokens = 0
		b.last = now
		return false
	}

	dur := now.Sub(b.last)
	cnt := int64(dur / b.rate) // 需要增加的次数

	b.last = now
	b.tokens += cnt
	if b.tokens > b.capacity {
		b.tokens = b.capacity
	}

	b.tokens -= n
	return true
}

func (b *Bucket) resetSize() int64 {
	size := (b.capacity - b.tokens) * int64(b.rate.Seconds())
	return time.Now().Unix() + size
}

// SetHeader 根据当前 Bucket 的情况设置报头
func (b *Bucket) SetHeader(w http.ResponseWriter) {
	w.Header().Set("X_rate-Limit-Limit", strconv.FormatInt(b.capacity, 10))
	w.Header().Set("X_rate-Limit-Remaining", strconv.FormatInt(b.tokens, 10))
	w.Header().Set("X_rate-Limit-Reset", strconv.FormatInt(b.resetSize(), 10))
}
