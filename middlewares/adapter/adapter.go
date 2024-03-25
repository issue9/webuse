// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

// Package adapter 与标准库的适配
package adapter

import (
	"net/http"

	"github.com/issue9/web"
)

// HTTPHandler 将 [http.Handler] 转换为 [web.HandlerFunc]
func HTTPHandler(h http.Handler) web.HandlerFunc {
	return func(ctx *web.Context) web.Responser {
		h.ServeHTTP(ctx, ctx.Request())
		return nil
	}
}

// HTTPHandlerFunc 将 [http.HandlerFunc] 转换为 [web.HandlerFunc]
func HTTPHandlerFunc(f func(http.ResponseWriter, *http.Request)) web.HandlerFunc {
	return HTTPHandler(http.HandlerFunc(f))
}
