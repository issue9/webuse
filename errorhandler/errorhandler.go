// SPDX-License-Identifier: MIT

// Package errorhandler 提供自定义错误页面的功能
package errorhandler

import (
	"net/http"

	"github.com/issue9/middleware/v4/recovery"
)

// HandleFunc 错误处理函数
//
// status 表示状态码，必须在第一时间输出；
type HandleFunc func(w http.ResponseWriter, status int)

// ErrorHandler 错误页面处理函数管理
//
// NOTE: 外层必须包含由 ErrorHandler.Recovery 声明的 recovery 中间件。
// 一旦写入由 ErrorHandler 托管的状态码，会直接中间整个中间件链的执行以 panic 的形式退出，
// 直接被 recovery 捕获。
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
func (e *ErrorHandler) Render(w http.ResponseWriter, status int) {
	f, found := e.handlers[status]
	if !found {
		f = defaultRender
	}

	f(w, status)
}

// Middleware 将当前中间件应用于 next
func (e *ErrorHandler) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(&response{ResponseWriter: w, eh: e}, r)
	})
}

// MiddlewareFunc 将当前中间件应用于 next
func (e *ErrorHandler) MiddlewareFunc(next func(http.ResponseWriter, *http.Request)) http.Handler {
	return e.Middleware(http.HandlerFunc(next))
}

// Recovery 生成一个可正确处理错误页面的 recovery.RecoverFunc 函数
//
// NOTE: ErrorHandler 最终是以特定的 panic 形式退出当前处理进程的，
// 所以必须要有 recover 函数捕获该 panic，否则会导致整个程序直接退出。
// 我们采用与 recovery 相结合的形式处理 panic，所以在 ErrorHandler
// 的外层必须要有一个由 ErrorHandler.Recovery 声明的 recovery.RecoverFunc 中间件。
func (e *ErrorHandler) Recovery(rf recovery.RecoverFunc) recovery.RecoverFunc {
	if rf == nil {
		rf = recovery.DefaultRecoverFunc(http.StatusInternalServerError)
	}

	return func(w http.ResponseWriter, msg interface{}) {
		// 通 httpStatus 退出的，并不能算是 panic，所以此处不输出调用堆栈信息。
		if status, ok := msg.(httpStatus); ok {
			if status > 0 {
				e.Render(w, int(status))
			}
			return
		}

		// 非 httpStatus 的退出，打印相关的错误信息到日志
		rf(w, msg)
	}
}

// 仅向客户端输出状态码
func defaultRender(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}
