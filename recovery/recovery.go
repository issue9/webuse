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
	"os"

	"github.com/issue9/mux/v5"
)

// RecoverFunc 错误处理函数
//
// msg 为输出的错误信息，可能是任意类型的数据，一般为从 recover() 返回的数据。
//
// NOTE: 并不能保证 w 是空白的，可能有内容已经输出，所以有关报头的操作可能会不启作用。
//
// github.com/issue9/mux/v5 用户可以直接使用其 mux.Recovery 处理会更好。
type RecoverFunc mux.RecoverFunc

// DefaultRecover RecoverFunc 的默认实现
//
// 向客户端输出 status 状态码，忽略其它任何信息。
func DefaultRecover(status int) RecoverFunc {
	return func(w http.ResponseWriter, msg interface{}) {
		http.Error(w, http.StatusText(status), status)
	}
}

// ConsoleRecover 向控制台输出错误信息
//
// status 表示向客户端输出的状态码。
func ConsoleRecover(status int) RecoverFunc {
	return func(w http.ResponseWriter, msg interface{}) {
		http.Error(w, http.StatusText(status), status)
		if _, err := fmt.Fprint(os.Stderr, msg); err != nil {
			panic(err)
		}
	}
}

// LogRecover 将错误信息输出到日志
//
// l 为输出的日志；status 表示向客户端输出的状态码。
func LogRecover(l *log.Logger, status int) RecoverFunc {
	return func(w http.ResponseWriter, msg interface{}) {
		http.Error(w, http.StatusText(status), status)
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
