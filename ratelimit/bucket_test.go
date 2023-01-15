// SPDX-License-Identifier: MIT

package ratelimit

import (
	"testing"
	"time"

	"github.com/issue9/assert/v3"
	"github.com/issue9/web"
	"github.com/issue9/web/server/servertest"
)

func TestBucket_allow(t *testing.T) {
	a := assert.New(t, false)
	s := servertest.NewTester(a, nil)
	// 由 gen 方法限定在同一个请求
	rate := New(s.Server(), "rl", 5, 10*time.Microsecond, func(*web.Context) (string, error) { return "1", nil })
	a.NotNil(rate)

	// NOTE: 根据机器配置，部分测试可能失败
	b := rate.newBucket()
	a.NotNil(b)
	a.False(b.allow(rate, 100)) // 超过最大量
	a.True(b.allow(rate, 1))
	a.True(b.allow(rate, 1))
	a.False(b.allow(rate, 4)) // 数量不够
	a.True(b.allow(rate, 3))  // 刚好拿完
	a.False(b.allow(rate, 1))

	time.Sleep(25 * time.Microsecond)
	a.True(b.allow(rate, 1))
	a.False(b.allow(rate, 5), "tokens=%v", b.Tokens)
}
