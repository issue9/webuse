// SPDX-License-Identifier: MIT

// Package errorhandler 提供自定义错误处理功能
//
// net/http 包中对于错误的处理是通过 http.Error() 进行的，
// 我们无法直接修改该方法，实现自定义的错误处理功能。
// 只能对 http.ResponseWriter.WriteHeader() 进行自定义，
// 在指定的状态下，抛出异常，再通过 recover 实现错误处理。
//
// 需要注意的是，如果采用了当前包的方案，那么默认情况下，
// 所有大于 400 的 WriteHeader 操作，都会被 panic，
// 如果你对某些操作不想按正常流程处理，可以使用 errorhandler.WriteHeader
// 代替默认的 ResponseWriter.WriteHeader 操作。
package errorhandler

import (
	"net/http"

	"github.com/issue9/middleware/v2/recovery"
)

// HandleFunc 错误处理函数，对某一固定的状态码可以做专门的处理
type HandleFunc func(http.ResponseWriter, int)

// ErrorHandler 错误处理函数的管理
type ErrorHandler struct {
	handlers map[int]HandleFunc
}

// New 声明 ErrorHandler 变量
func New() *ErrorHandler {
	return &ErrorHandler{
		handlers: make(map[int]HandleFunc, 20),
	}
}

// Add 添加针对指定状态码的错误处理函数
func (e *ErrorHandler) Add(f HandleFunc, status ...int) (ok bool) {
	for _, s := range status {
		if _, found := e.handlers[s]; found {
			return false
		}

		e.handlers[s] = f
	}

	return true
}

// Set 添加或修改指定状态码对应的处理函数
//
// 有则修改，没有则添加，如果 f 为 nil，则表示删除该状态码的处理函数。
//
// status 表示处理函数 f 对应的状态码，仅对大于等于 400 的启作用，
// 同时还有一个特殊的状态码 0，表示那些未设置的状态码会统一采和此处理函数。
// 如果也没设置 0，则仅简单地输出状态码对应的错误信息。
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
		if f, found = e.handlers[0]; !found || f == nil {
			f = defaultRender
		}
	} else if f == nil {
		f = defaultRender
	}

	f(w, status)
}

// Middleware 将当前中间件应用于 next
//
// NOTE: 要求在最外层
func (e *ErrorHandler) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(&response{ResponseWriter: w}, r)
	})
}

// MiddlewareFunc 将当前中间件应用于 next
//
// NOTE: 要求在最外层
func (e *ErrorHandler) MiddlewareFunc(next func(http.ResponseWriter, *http.Request)) http.Handler {
	return e.Middleware(http.HandlerFunc(next))
}

// Recovery 生成一个 recovery.RecoverFunc 函数用于捕获由 panic 触发的事件
//
// 相较于 recovery 的相关功能，此函数可以正常处理 errorhandler 的错误代码。
// rf 表示在处理完 errorhandler 的相关功能之后，后续的处理方式，如果为空则采用
// recovery.DefaultRecoverFunc(500)。
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
