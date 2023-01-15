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
	Tokens int       `json:"tokens"` // 可用令牌数量
	Last   time.Time `json:"last"`   // 最后次添加令牌的时间
}

// 生成一个新的令牌桶，新的令牌桶中令牌数量为满格
func (rate *Ratelimit) newBucket() *Bucket {
	return &Bucket{
		Tokens: rate.capacity, // 默认为满格
		Last:   time.Now(),
	}
}

// 是否允许拿走 n 块令牌
func (b *Bucket) allow(rate *Ratelimit, n int) bool {
	if n > rate.capacity { // 超过最大值
		return false
	}

	now := time.Now()
	dur := now.Sub(b.Last)      // 从上次拿令牌到现在的时间段
	cnt := int(dur / rate.rate) // 计算这段时间内需要增加的令牌
	b.Tokens += cnt
	if b.Tokens > rate.capacity {
		b.Tokens = rate.capacity
	}

	if b.Tokens < n { // 不够
		return false
	}

	b.Last = now
	b.Tokens -= n
	return true
}

func (b *Bucket) setHeader(rate *Ratelimit, ctx *web.Context) {
	t := (rate.capacity - b.Tokens) * rate.rateSeconds
	rest := time.Now().Unix() + int64(t)

	h := ctx.Header()
	h.Set("X-Rate-Limit-Limit", strconv.Itoa(rate.capacity))
	h.Set("X-Rate-Limit-Remaining", strconv.Itoa(b.Tokens))
	h.Set("X-Rate-Limit-Reset", strconv.FormatInt(rest, 10))
}
