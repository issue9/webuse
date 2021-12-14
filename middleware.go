// SPDX-License-Identifier: MIT

// Package middleware 包含了一系列 http.Handler 接口的中间件
package middleware

import (
	"net/http"

	"github.com/issue9/mux/v5/middleware"
)

type (
	// Middlewares 中间件管理
	Middlewares = middleware.Middlewares

	MiddlewareFunc = middleware.Func
)

// NewMiddlewares 声明新的 Middlewares 实例
func NewMiddlewares(next http.Handler) *Middlewares {
	return middleware.NewMiddlewares(next)
}
