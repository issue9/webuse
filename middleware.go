// SPDX-License-Identifier: MIT

// Package middleware 包含了一系列 http.Handler 接口的中间件
package middleware

import (
	"net/http"

	"github.com/issue9/mux/v5"
)

// Middlewares 中间件管理
type Middlewares = mux.Middlewares

// NewMiddlewares 声明新的 Middlewares 实例
func NewMiddlewares(next http.Handler) *Middlewares {
	return mux.NewMiddlewares(next)
}
