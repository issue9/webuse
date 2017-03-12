// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"fmt"
	"net/http"
	"runtime"
)

// RecoverFunc 错误处理函数。Recovery 需要此函数作为出错时的处理。
//
// msg 为输出的错误信息，可能是任意类型的数据，一般为从 recover() 返回的数据。
type RecoverFunc func(w http.ResponseWriter, msg interface{})

// RecoverFunc 的默认实现。
//
// 为一个简单的 500 错误信息。不会输出 msg 参数的内容。
func defaultRecoverFunc(w http.ResponseWriter, msg interface{}) {
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// PrintDebug 是 RecoverFunc 类型的实现。方便 NewRecovery 在调度期间将函数的调用信息输出到 w。
func PrintDebug(w http.ResponseWriter, msg interface{}) {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintln(w, msg)
	for i := 1; true; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			return
		}

		fmt.Fprintf(w, "@ %v:%v\n", file, line)
	}
}

// 捕获并处理 panic 信息。
type recovery struct {
	handler     http.Handler
	recoverFunc RecoverFunc
}

// Recovery 用于处理当发生 panic 时的处理的 handler。
// h 参数中发生的panic将被截获并处理，不会再向上级反映。
//
// Recovery 应该处在所有 http.Handler 的最外层，用于处理所有没有被处理的 panic。
//
// 当 h 参数为空时，将直接 panic。
// rf 参数用于指定处理 panic 信息的函数，其原型为 RecoverFunc，
// 当将 rf 指定为 nil 时，将使用默认的处理函数，仅仅向客户端输出 500 的错误信息，没有具体内容。
func Recovery(h http.Handler, rf RecoverFunc) *recovery {
	if h == nil {
		panic("handlers.Recovery:参数h不能为空")
	}

	if rf == nil {
		rf = defaultRecoverFunc
	}

	return &recovery{
		handler:     h,
		recoverFunc: rf,
	}
}

// RecoveryFunc 将一个 http.HandlerFunc 包装成 http.Handler
func RecoveryFunc(f func(http.ResponseWriter, *http.Request), rf RecoverFunc) *recovery {
	return Recovery(http.HandlerFunc(f), rf)
}

// implement net/http.Handler.ServeHTTP(...)
func (r *recovery) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			r.recoverFunc(w, err)
		}
	}()

	r.handler.ServeHTTP(w, req)
}
