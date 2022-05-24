// SPDX-License-Identifier: MIT

// Package access 记录接口访问记录
package access

import (
	"github.com/issue9/logs/v4"
	"github.com/issue9/web"
	"github.com/issue9/web/server"
)

// Access 生成用于打印访问记录的中间件
func Access(level logs.Level) web.Middleware {
	return web.MiddlewareFunc(func(next web.HandlerFunc) web.HandlerFunc {
		return func(ctx *server.Context) server.Responser {
			ctx.OnExit(func(status int) {
				r := ctx.Request()
				ctx.Server().Logs().Logger(level).Printf("[%d] %s\t%s\n", status, r.Method, r.URL.String())
			})
			return next(ctx)
		}
	})
}
