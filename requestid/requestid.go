// SPDX-License-Identifier: MIT

// Package requestid
package requestid

import (
	"github.com/issue9/unique"
	"github.com/issue9/web"
)

const varValue varType = 1

type varType int

// Gen 生成唯一 ID 的方法
type Gen func() string

type requestID struct {
	key string
	gen Gen
}

// New 声明 request id 的中间件
//
// key 表示报头名称，如果为空，则采用 x-request-id；
// gen 为生成 id 的方法，可以为空，采用默认值；
func New(key string, gen Gen) web.Middleware {
	if key == "" {
		key = "X-Request-ID"
	}

	if gen == nil {
		gen = func() string { return unique.String().String() }
	}

	return &requestID{key: key, gen: gen}
}

func (r *requestID) Middleware(next web.HandlerFunc) web.HandlerFunc {
	return func(ctx *web.Context) web.Responser {
		id := ctx.Request().Header.Get(r.key)
		if id == "" {
			id = r.gen()
		}

		ctx.Vars[varValue] = id
		ctx.Header().Set(r.key, id)
		return next(ctx)
	}
}

// Get 从 [server.Context] 中获取 request id 值
func Get(ctx *web.Context) string {
	if v, found := ctx.Vars[varValue]; found {
		return v.(string)
	}
	panic("并未设置 request id 中间件")
}
