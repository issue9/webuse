// SPDX-License-Identifier: MIT

package ratelimit

import (
	"strconv"
	"time"

	"github.com/issue9/web"
)

// Bucket 令牌桶
//
// 真正的令牌桶算法应该是按时自动增加令牌数量，
// 此处稍作修改：去掉自动分发令牌功能，集中在每次拿令牌时，
// 一次性补全之前缺的令牌数量。
type Bucket struct {
	Capacity int64         `json:"cap"`    // 上限
	Rate     time.Duration `json:"rate"`   // 每隔 rate 添加一个令牌
	Tokens   int64         `json:"tokens"` // 可用令牌数量
	Last     time.Time     `json:"last"`   // 最后次添加令牌的时间
}

// 生成一个新的令牌桶，新的令牌桶中令牌数量为满格
//
// capacity 令牌桶中最大的令牌数量。
// rate 产生令牌的速度。
func newBucket(capacity int64, rate time.Duration) *Bucket {
	return &Bucket{
		Capacity: capacity,
		Rate:     rate,
		Tokens:   capacity, // 默认为满格
		Last:     time.Now(),
	}
}

// 是否允许拿走 n 块令牌。
func (b *Bucket) allow(n int64) bool {
	if n > b.Capacity { // 超过最大值
		return false
	}

	now := time.Now()
	dur := now.Sub(b.Last)     // 从上次拿令牌到现在的时间
	cnt := int64(dur / b.Rate) // 计算这段时间内需要增加的令牌
	b.Tokens += cnt
	if b.Tokens > b.Capacity {
		b.Tokens = b.Capacity
	}

	if b.Tokens < n { // 不够
		return false
	}

	b.Last = now
	b.Tokens -= n
	return true
}

// 获取 X-Rate-Limit-Reset 的值
func (b *Bucket) resetTime() int64 {
	t := (b.Capacity - b.Tokens) * int64(b.Rate.Seconds())
	return time.Now().Unix() + t
}

func (b *Bucket) setHeader(ctx *web.Context) {
	h := ctx.Header()
	h.Set("X-Rate-Limit-Limit", strconv.FormatInt(b.Capacity, 10))
	h.Set("X-Rate-Limit-Remaining", strconv.FormatInt(b.Tokens, 10))
	h.Set("X-Rate-Limit-Reset", strconv.FormatInt(b.resetTime(), 10))
}
