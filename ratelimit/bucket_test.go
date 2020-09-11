// SPDX-License-Identifier: MIT

package ratelimit

import (
	"testing"
	"time"

	"github.com/issue9/assert"
)

func TestBucket_allow(t *testing.T) {
	a := assert.New(t)

	// NOTE: 根据机器配置，部分测试可能失败
	b := newBucket(5, 10*time.Microsecond)
	a.NotNil(b)
	a.False(b.allow(100)) // 超过最大量
	a.True(b.allow(1))
	a.True(b.allow(1))
	a.False(b.allow(4)) // 数量不够
	a.True(b.allow(3))  // 刚好拿完
	a.False(b.allow(1))

	time.Sleep(10 * time.Microsecond)
	a.True(b.allow(1))
	a.False(b.allow(5), "tokens=%v", b.Tokens)
}
