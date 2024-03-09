// SPDX-FileCopyrightText: 2015-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package access 记录接口访问记录
package access

import "github.com/issue9/web"

// New 声明 [Access] 中间件
//
// l 表示记录输出的通道；
// format 表示记录的格式，接受三个参数，分别为状态码、请求方法和请求地址，
// 默认使用 "[%d] %s\t%s\n"，可以使用 [fmt.Sprintf] 的顺序标记作调整；
func New(l *web.Logger, format string) web.Plugin {
	const defaultFormat = "[%d] %s\t%s\n"
	if format == "" {
		format = defaultFormat
	}

	return web.PluginFunc(func(s web.Server) {
		s.OnExitContext(func(ctx *web.Context, status int) {
			r := ctx.Request()
			l.Printf(format, status, r.Method, r.URL.String())
		})
	})
}
