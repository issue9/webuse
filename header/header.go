// SPDX-License-Identifier: MIT

// Package header 用于指定输出的报头
package header

import "net/http"

type header struct {
	headers     map[string]string // 静态内容
	headersFunc func(http.Header) // 动态生成的内容
	handler     http.Handler
}

// New 声明一个用于输出报头的中间件
//
// 如果 funcs 不为空，则 funcs 与 headers 相同的内容，
// 以 funcs 为最终内容。
func New(next http.Handler, headers map[string]string, funs func(http.Header)) http.Handler {
	return &header{
		handler:     next,
		headers:     headers,
		headersFunc: funs,
	}
}

func (h *header) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for k, v := range h.headers {
		w.Header().Set(k, v)
	}

	if h.headersFunc != nil {
		h.headersFunc(w.Header())
	}

	h.handler.ServeHTTP(w, r)
}
