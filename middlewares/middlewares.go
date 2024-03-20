// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

// Package middlewares 适用于 [web.Middleware] 的中间件
package middlewares

import "github.com/issue9/web"

// Plugin 将中间件 m 转换为插件
//
// 返回的插件会将中间件 m 作用于所有的路由。
func Plugin(m web.Middleware) web.Plugin {
	return web.PluginFunc(func(s web.Server) { s.Routers().Use(m) })
}
