// SPDX-License-Identifier: MIT

// Package errorhandler 提供自定义错误页面的功能
package errorhandler

import (
	"io"
	"net/http"

	"github.com/issue9/middleware/v5"
)

// HandleFunc 错误处理函数
type HandleFunc func(w io.Writer, status int)

// ErrorHandler 错误页面处理函数管理
//
// 一旦写入由 ErrorHandler 托管的状态码，会直接中间整个中间件链的执行以 panic 的形式退出。
type ErrorHandler struct {
	handlers map[int]HandleFunc
}

// New 声明 ErrorHandler 变量
func New() *ErrorHandler {
	return &ErrorHandler{handlers: make(map[int]HandleFunc, 20)}
}

// Add 添加针对指定状态码的错误处理函数
//
// NOTE: 如果指定了 400 以下的状态码，那么该状态码也会被当作错误页面进行托管。
func (e *ErrorHandler) Add(f HandleFunc, status ...int) (ok bool) {
	if f == nil {
		panic("参数 f 不能为 nil")
	}

	for _, s := range status {
		if _, found := e.handlers[s]; found {
			return false
		}

		e.handlers[s] = f
	}

	return true
}

// Exists 指定状态码对应的处理函数是否存在
func (e *ErrorHandler) Exists(status int) bool {
	_, exists := e.handlers[status]
	return exists
}

// Set 添加或修改指定状态码对应的处理函数
//
// 有则修改，没有则添加，如果 f 为 nil，则表示删除该状态码的处理函数。
//
// NOTE: 如果指定了 400 以下的状态码，那么该状态码也会被当作错误页面进行托管。
func (e *ErrorHandler) Set(f HandleFunc, status ...int) {
	if f == nil {
		for _, s := range status {
			delete(e.handlers, s)
		}
		return
	}

	for _, s := range status {
		e.handlers[s] = f
	}
}

// Render 向客户端输出指定状态码的错误内容
func (e *ErrorHandler) Render(w io.Writer, status int) {
	f, found := e.handlers[status]
	if !found {
		f = defaultRender
	}
	f(w, status)
}

func (e *ErrorHandler) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(&response{ResponseWriter: w, eh: e}, r)
	})
}

func (e *ErrorHandler) MiddlewareFunc(next func(http.ResponseWriter, *http.Request)) http.Handler {
	return e.Middleware(http.HandlerFunc(next))
}

// Renderer 返回一个用于输出错误页面的中间件
//
// 此中间件应该在 ErrorHandler 中间件的外层调用，
// 只有这样 ErrorHandler 抛出的代码才能被捕获。
func (e *ErrorHandler) Renderer() middleware.MiddlewareFunc {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if msg := recover(); msg != nil {
					if status, ok := msg.(httpStatus); ok {
						e.Render(w, int(status))
					}
				}
			}()

			h.ServeHTTP(w, r)
		})
	}
}

// 仅向客户端输出状态码
func defaultRender(w io.Writer, status int) {
	io.WriteString(w, http.StatusText(status))
}
