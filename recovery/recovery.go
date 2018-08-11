// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package recovery 提供了处理 panic 操作的中间件。
package recovery

import (
	"fmt"
	"net/http"

	"github.com/issue9/utils"
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

// PrintDebug 是 RecoverFunc 类型的实现。方便 NewRecovery 在调试期间将函数的调用信息输出到 w。
func PrintDebug(w http.ResponseWriter, msg interface{}) {
	w.WriteHeader(http.StatusNotFound)

	data, err := utils.TraceStack(2, msg)
	if err != nil {
		panic(err)
	}

	if _, err = fmt.Fprint(w, data); err != nil {
		panic(err)
	}
}

// New 声明一个处理 panic 操作的中间件。
// next 参数中发生的 panic 将被截获并处理，不会再向上级反映。
//
// 当 next 参数为空时，将直接 panic。
// rf 参数用于指定处理 panic 信息的函数，其原型为 RecoverFunc，
// 当将 rf 指定为 nil 时，将使用默认的处理函数，仅仅向客户端输出 500 的错误信息，没有具体内容。
func New(next http.Handler, rf RecoverFunc) http.Handler {
	if next == nil {
		panic("参数 h 不能为空")
	}

	if rf == nil {
		rf = defaultRecoverFunc
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				rf(w, err)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
