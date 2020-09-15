// SPDX-License-Identifier: MIT

// Package header 用于指定输出的报头
package header

import "net/http"

// Header 修正报头输出内容
type Header struct {
	headers     map[string]string // 静态内容
	headersFunc func(http.Header) // 动态生成的内容
}

// New 声明一个用于输出报头的中间件
//
// 如果 funcs 不为空，则 funcs 与 headers 相同的内容，
// 以 funcs 为最终内容。
func New(headers map[string]string, funcs func(http.Header)) *Header {
	return &Header{
		headers:     headers,
		headersFunc: funcs,
	}
}

// Set 添加或是修改报头
func (h *Header) Set(name, value string) *Header {
	h.headers[name] = value
	return h
}

// Delete 删除指定的报头
func (h *Header) Delete(name string) *Header {
	delete(h.headers, name)
	return h
}

// Middleware 将当前中间件应用于 next
func (h *Header) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for k, v := range h.headers {
			w.Header().Set(k, v)
		}

		if h.headersFunc != nil {
			h.headersFunc(w.Header())
		}

		next.ServeHTTP(w, r)
	})
}

// MiddlewareFunc 将当前中间件应用于 next
func (h *Header) MiddlewareFunc(next func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return h.Middleware(http.HandlerFunc(next))
}
