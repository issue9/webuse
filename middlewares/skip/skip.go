// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

// Package skip 是否根据条件跳过路由的执行
package skip

import "github.com/issue9/web"

// New 只有在 cond 为 true 时才执行路由
//
// problemID 表示在条件为 false 时返回的错误码。
func New(cond func(*web.Context) bool, problemID string) web.MiddlewareFunc {
	return web.MiddlewareFunc(func(next web.HandlerFunc, _, _, _ string) web.HandlerFunc {
		return func(ctx *web.Context) web.Responser {
			if cond(ctx) {
				return next(ctx)
			}
			return ctx.Problem(problemID)
		}
	})
}
