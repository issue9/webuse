// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package debug 提供调试相关功能
package debug

import (
	"expvar"
	"net/http/pprof"
	"strings"

	"github.com/issue9/mux/v8/header"
	"github.com/issue9/web"
)

// New 输出调试信息
//
// p 是指路由中的参数名，比如以下示例中，p 的值为 debug：
//
//	r.Get("/test{debug}", New("debug", web.ProblemNotFound))
//
// p 所代表的路径包含了前缀的 /。
func New(name, nameNotExists string) web.HandlerFunc {
	return func(ctx *web.Context) web.Responser {
		p, resp := ctx.PathString(name, nameNotExists)
		if resp != nil {
			return resp
		}

		switch {
		case p == "/vars":
			expvar.Handler().ServeHTTP(ctx, ctx.Request())
		case p == "/pprof/cmdline":
			pprof.Cmdline(ctx, ctx.Request())
		case p == "/pprof/profile":
			pprof.Profile(ctx, ctx.Request())
		case p == "/pprof/symbol":
			pprof.Symbol(ctx, ctx.Request())
		case p == "/pprof/trace":
			pprof.Trace(ctx, ctx.Request())
		case p == "/pprof/goroutine":
			pprof.Handler("goroutine").ServeHTTP(ctx, ctx.Request())
		case p == "/pprof/threadcreate":
			pprof.Handler("threadcreate").ServeHTTP(ctx, ctx.Request())
		case p == "/pprof/mutex":
			pprof.Handler("mutex").ServeHTTP(ctx, ctx.Request())
		case p == "/pprof/heap":
			pprof.Handler("heap").ServeHTTP(ctx, ctx.Request())
		case p == "/pprof/block":
			pprof.Handler("block").ServeHTTP(ctx, ctx.Request())
		case p == "/pprof/allocs":
			pprof.Handler("allocs").ServeHTTP(ctx, ctx.Request())
		case strings.HasPrefix(p, "/pprof/"):
			// pprof.Index 写死了 /debug/pprof，所以直接替换这个变量
			r := ctx.Request()
			r.URL.Path = "/debug/pprof/" + strings.TrimPrefix(p, "/pprof/")
			pprof.Index(ctx, r)
		case p == "/":
			// ctx.SetMimetype(header.HTML) SetMimetype 需要在 Codec 中已经正确设置
			ctx.Header().Set(header.ContentType, header.HTML)
			if _, err := ctx.Write(debugHtml); err != nil {
				ctx.Logs().ERROR().Error(err)
			}
			return nil
		default:
			return ctx.NotFound()
		}
		return nil
	}
}

var debugHtml = []byte(`
<!DOCTYPE HTML>
<html>
	<head>
		<title>Debug</title>
		<meta charset="utf-8" />
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
	</head>
	<body>
		<a href="vars">vars</a><br />
		<a href="pprof/cmdline">pprof/cmdline</a><br />
		<a href="pprof/profile">pprof/profile</a><br />
		<a href="pprof/symbol">pprof/symbol</a><br />
		<a href="pprof/trace">pprof/trace</a><br />
		<a href="pprof/goroutine">pprof/goroutine</a><br />
		<a href="pprof/threadcreate">pprof/threadcreate</a><br />
		<a href="pprof/mutex">pprof/mutex</a><br />
		<a href="pprof/heap">pprof/heap</a><br />
		<a href="pprof/block">pprof/block</a><br />
		<a href="pprof/allocs">pprof/allocs</a><br />
		<a href="pprof/">pprof/</a>
	</body>
</html>
`)
