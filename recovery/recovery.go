// SPDX-License-Identifier: MIT

// Package recovery 提供了处理 panic 操作的中间件
//
//  recovery.RecoverFunc(func(w http.ResponseWriter, msg interface{}) {
//      fmt.Printf("recovery: %s", msg)
//  }).MiddlewareFunc(func(w http.ResponseWriter, r *http.Request){
//      panic("panic")
//  })
package recovery

import (
	"fmt"
	"log"
	"net/http"

	"github.com/issue9/source"
)

// RecoverFunc 错误处理函数
//
// msg 为输出的错误信息，可能是任意类型的数据，一般为从 recover() 返回的数据。
type RecoverFunc func(w http.ResponseWriter, msg interface{})

// DefaultRecoverFunc RecoverFunc 的默认实现
//
// 为一个简单的状态码信息输出。不会输出 msg 参数的内容。
func DefaultRecoverFunc(status int) RecoverFunc {
	return func(w http.ResponseWriter, msg interface{}) {
		http.Error(w, http.StatusText(status), status)
	}
}

// TraceStack 打印调用的堆栈信息的 RecoverFunc 实现
//
// 调用堆栈信息将输出到客户端。
func TraceStack(status int) RecoverFunc {
	return func(w http.ResponseWriter, msg interface{}) {
		w.WriteHeader(status)

		data, err := source.TraceStack(2, msg)
		if err != nil {
			panic(err)
		}

		if _, err = fmt.Fprint(w, data); err != nil {
			panic(err)
		}
	}
}

// LogTraceStack 打印调用信息到日志的 RecoverFunc 实现
func LogTraceStack(l *log.Logger, status int) RecoverFunc {
	return func(w http.ResponseWriter, msg interface{}) {
		w.WriteHeader(status)
		l.Println(msg)
	}
}

// MiddlewareFunc 将当前中间件应用于 next
func (rf RecoverFunc) MiddlewareFunc(next func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return rf.Middleware(http.HandlerFunc(next))
}

// Middleware 将当前中间件应用于 next
func (rf RecoverFunc) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				rf(w, err)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
