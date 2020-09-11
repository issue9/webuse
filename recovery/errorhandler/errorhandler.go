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
	"fmt"
	"log"
	"net/http"

	"github.com/issue9/source"

	"github.com/issue9/middleware/recovery"
)

// HandleFunc 错误处理函数，对某一固定的状态码可以做专门的处理
type HandleFunc func(http.ResponseWriter, int)

// ErrorHandler 错误处理函数的管理
type ErrorHandler struct {
	// 指定状态下对应的错误处理函数。
	//
	// 若该状态码的处理函数不存在，则会查找键值为 0 的函数，
	// 若依然不存在，则调用 defaultRender
	//
	// 用户也可以通过调用 Add 进行添加。
	handlers map[int]HandleFunc
}

// New 声明 ErrorHandler 变量
func New() *ErrorHandler {
	return &ErrorHandler{
		handlers: make(map[int]HandleFunc, 20),
	}
}

// Add 添加针对特写状态码的错误处理函数
func (e *ErrorHandler) Add(f HandleFunc, status ...int) error {
	for _, s := range status {
		if _, found := e.handlers[s]; found {
			return fmt.Errorf("状态码 %d 已经存在", s)
		}

		e.handlers[s] = f
	}

	return nil
}

// Set 设置指定状态码对应的处理函数
//
// 有则修改，没有则添加
func (e *ErrorHandler) Set(f HandleFunc, status ...int) {
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

// New 构建一个可以捕获错误状态码的 Handler
//
// NOTE: 要求在最外层
func (e *ErrorHandler) New(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(&response{ResponseWriter: w}, r)
	})
}

// NewFunc 构建一个可以捕获错误状态码的 Handler
//
// NOTE: 要求在最外层
func (e *ErrorHandler) NewFunc(next func(http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next(&response{ResponseWriter: w}, r)
	})
}

// Recovery 生成一个 recovery.RecoverFunc 函数用于捕获由 panic 触发的事件
//
// errlog 表示输出调用堆栈信息到日志。可以为空，表示不输出信息。
func (e *ErrorHandler) Recovery(errlog *log.Logger) recovery.RecoverFunc {
	return func(w http.ResponseWriter, msg interface{}) {
		// 通 httpStatus 退出的，并不能算是错误，所以此处不输出调用堆栈信息。
		if status, ok := msg.(httpStatus); ok {
			if status > 0 {
				e.Render(w, int(status))
			}
			return
		}

		// 非 httpStatus 的退出，打印相关的错误信息到日志

		e.Render(w, http.StatusInternalServerError)

		if errlog != nil {
			message, err := source.TraceStack(3, msg)
			if err != nil {
				panic(err)
			}

			errlog.Println(message)
		}
	}
}

// 仅向客户端输出状态码
func defaultRender(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}
