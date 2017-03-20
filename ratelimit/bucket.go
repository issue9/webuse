// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package ratelimit

import (
	"net/http"
	"strconv"
	"time"
)

// Bucket 令牌桶。
//
// 真正的令牌桶算法应该是按时自动增加令牌数量，
// 此处稍作修改：去掉自动分发令牌功能，集中在次拿令牌时，
// 一次性补全之前缺的令牌数量。
type Bucket struct {
	capacity int64         // 上限
	rate     time.Duration // 每隔 rate 添加一个令牌
	tokens   int64         // 可用令牌数量
	last     time.Time     // 最后次添加令牌的时间
}

// 生成一个新的令牌桶，新的令牌桶中令牌数量为满格。
//
// capacity 令牌桶中最大的令牌数量。
// rate 产生令牌的速度。
func newBucket(capacity int64, rate time.Duration) *Bucket {
	return &Bucket{
		capacity: capacity,
		rate:     rate,
		tokens:   capacity, // 默认为满格
		last:     time.Now(),
	}
}

// 是否允许拿走 n 块令牌。
func (b *Bucket) allow(n int64) bool {
	now := time.Now()

	if n > b.capacity {
		b.tokens = 0
		b.last = now
		return false
	}

	dur := now.Sub(b.last)     // 从上次存储到现在的时间
	cnt := int64(dur / b.rate) // 需要增加的次数

	b.last = now
	b.tokens += cnt
	if b.tokens > b.capacity {
		b.tokens = b.capacity
	}

	b.tokens -= n
	return true
}

// 获取 x-rate-reset 的值。
func (b *Bucket) resetSize() int64 {
	size := (b.capacity - b.tokens) * int64(b.rate.Seconds())
	return time.Now().Unix() + size
}

// SetHeader 根据当前 Bucket 的情况设置报头
func (b *Bucket) SetHeader(w http.ResponseWriter) {
	w.Header().Set("X-Rate-Limit-Limit", strconv.FormatInt(b.capacity, 10))
	w.Header().Set("X-Rate-Limit-Remaining", strconv.FormatInt(b.tokens, 10))
	w.Header().Set("X-Rate-Limit-Reset", strconv.FormatInt(b.resetSize(), 10))
}
