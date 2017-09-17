// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package header 用于指定输出的报头。
package header

import "net/http"

type header struct {
	headers     map[string]string
	headersFunc map[string]func() string
	handler     http.Handler
}

// New 声明一个用于输出报头的中间件。
func New(next http.Handler, headers map[string]string, funs map[string]func() string) http.Handler {
	return &header{
		handler:     next,
		headers:     headers,
		headersFunc: funs,
	}
}

func (h *header) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if len(h.headers) > 0 {
		for k, v := range h.headers {
			w.Header().Set(k, v)
		}
	}

	if len(h.headersFunc) > 0 {
		for k, v := range h.headersFunc {
			w.Header().Set(k, v())
		}

	}

	h.handler.ServeHTTP(w, r)
}
