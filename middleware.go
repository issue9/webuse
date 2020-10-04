// SPDX-License-Identifier: MIT

// Package middleware 包含了一系列 http.Handler 接口的中间件
package middleware

import "net/http"

// Middleware 将一个 http.Handler 封装成另一个 http.Handler
type Middleware func(http.Handler) http.Handler

// Handler 按顺序将所有的中间件应用于 h
func Handler(h http.Handler, middleware ...Middleware) http.Handler {
	if l := len(middleware); l > 0 {
		for i := l - 1; i >= 0; i-- {
			h = middleware[i](h)
		}
	}
	return h
}

// HandlerFunc 按顺序将所有的中间件应用于 h
func HandlerFunc(h func(w http.ResponseWriter, r *http.Request), middleware ...Middleware) http.Handler {
	return Handler(http.HandlerFunc(h), middleware...)
}
