// SPDX-License-Identifier: MIT

// Package debugger 提供测试和性能测试相关的中间件
package debugger

import (
	"expvar"
	"net/http"
	"net/http/pprof"
	"strings"
)

// Debugger 用于调试的中间件
type Debugger struct {
	// Pprof 设置了 net/http/pprof 中与性能测试相关的页面地址前缀
	//
	// 如果此值为空，则表示不会启用这些功能。
	Pprof string

	// Vars 设置了 expvar 中相关的测试数据项地址
	//
	// 如果此值为空，则表示不会启用这些功能。
	Vars string
}

// Middleware 将当前中间件应用于 next
func (dbg *Debugger) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case dbg.Pprof != "" && strings.HasPrefix(r.URL.Path, dbg.Pprof):
			switch path := r.URL.Path[len(dbg.Pprof):]; path {
			case "cmdline":
				pprof.Cmdline(w, r)
			case "profile":
				pprof.Profile(w, r)
			case "symbol":
				pprof.Symbol(w, r)
			case "trace":
				pprof.Trace(w, r)
			default:
				r.URL.Path = "/debug/pprof/" + path // pprof.Index 写死了 /debug/pprof，所以直接替换这个变量
				pprof.Index(w, r)
			}
		case dbg.Vars != "" && strings.HasPrefix(r.URL.Path, dbg.Vars):
			expvar.Handler().ServeHTTP(w, r)
		default:
			next.ServeHTTP(w, r)
		}
	})
}

// MiddlewareFunc 将当前中间件应用于 next
func (dbg *Debugger) MiddlewareFunc(next func(http.ResponseWriter, *http.Request)) http.Handler {
	return dbg.Middleware(http.HandlerFunc(next))
}
