// SPDX-License-Identifier: MIT

// Package middleware 包含了一系列 http.Handler 接口的中间件
package middleware

import (
	"net/http"

	"github.com/issue9/mux/v4"
)

// Middlewares 中间件管理
type Middlewares struct {
	http.Handler
	middlewares []mux.MiddlewareFunc
	next        http.Handler
}

// NewMiddlewares 声明新的 Middlewares 实例
func NewMiddlewares(next http.Handler) *Middlewares {
	return &Middlewares{
		Handler:     next,
		middlewares: make([]mux.MiddlewareFunc, 0, 10),
		next:        next,
	}
}

// Prepend 添加中间件到顶部
//
// 顶部的中间件在运行过程中将最早被调用，多次添加，则最后一次的在顶部。
func (mgr *Middlewares) Prepend(m mux.MiddlewareFunc) *Middlewares {
	mgr.middlewares = append(mgr.middlewares, m)
	mgr.Handler = mux.ApplyMiddlewares(mgr.next, mgr.middlewares...)
	return mgr
}

// Append 添加中间件到尾部
//
// 尾部的中间件将最后被调用，多次添加，则最后一次的在最末尾。
func (mgr *Middlewares) Append(m mux.MiddlewareFunc) *Middlewares {
	ms := make([]mux.MiddlewareFunc, 0, 1+len(mgr.middlewares))
	ms = append(ms, m)
	if len(mgr.middlewares) > 0 {
		ms = append(ms, mgr.middlewares...)
	}
	mgr.middlewares = ms
	mgr.Handler = mux.ApplyMiddlewares(mgr.next, mgr.middlewares...)
	return mgr
}

// Reset 清除中间件
func (mgr *Middlewares) Reset() *Middlewares {
	mgr.middlewares = mgr.middlewares[:0]
	mgr.Handler = mgr.next
	return mgr
}
