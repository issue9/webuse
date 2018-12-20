// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package host 提供了限定访问域名的中间件。
package host

import (
	"net/http"
	"strings"
)

type host struct {
	domains []string
	handler http.Handler
}

// New 声明一个限定域名的中间件.
//
// 若请求的域名不允许，会返回 404 错误。
// 若 domains 为空，则任何请求都将返回 404。
//
// 仅会将域名与 domains 进行比较，端口与协议都将不参写比较。
func New(next http.Handler, domains ...string) http.Handler {
	return &host{
		domains: domains,
		handler: next,
	}
}

func (h *host) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// r.URL.Hostname() 可能是空值
	hostname := r.Host
	index := strings.IndexByte(hostname, ':')
	if index >= 0 {
		hostname = hostname[:index]
	}

	for _, domain := range h.domains {
		if domain == hostname {
			h.handler.ServeHTTP(w, r)
			return
		}
	}

	http.NotFound(w, r)
}
