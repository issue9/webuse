// SPDX-FileCopyrightText: 2025 caixw
//
// SPDX-License-Identifier: MIT

// package mimetype 提供为限定可用的媒体类型的中间件
package mimetype

import (
	"net/http"
	"slices"
	"strings"

	"github.com/issue9/mux/v9/header"
	"github.com/issue9/web"
)

// New 声明用于限定媒体类型的中间件
//
// 该中间件会限定接口的媒体的类型，对于 OPTIONS 方法，会设置 Accept 报头，
// 对于其它请求方法，会检查请求的媒体类型是否在指定的列表中。
func New(t ...string) web.MiddlewareFunc {
	return func(next web.HandlerFunc, method, pattern, router string) web.HandlerFunc {
		if method == http.MethodOptions {
			return func(ctx *web.Context) web.Responser {
				r := next(ctx)
				ctx.Header().Set(header.Accept, strings.Join(t, ","))
				return r
			}
		}

		return func(ctx *web.Context) web.Responser {
			if !slices.Contains(t, ctx.Mimetype(false)) {
				return ctx.Problem(web.ProblemNotAcceptable)
			}
			return next(ctx)
		}
	}
}
