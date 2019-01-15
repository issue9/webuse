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
	domains   []string // 域名列表
	wildcards []string // 泛域名列表，只保存 * 之后的部分内容
	handler   http.Handler
}

// New 声明一个限定域名的中间件.
//
// 若请求的域名不允许，会返回 404 错误。
// 若 domains 为空，则任何请求都将返回 404。
//
// 仅会将域名与 domains 进行比较，端口与协议都将不参写比较。
// domains 可以是泛域名，比如 *.example.com，但不能是 s1.*.example.com
func New(next http.Handler, domains ...string) http.Handler {
	return newHost(next, domains...)
}

func newHost(next http.Handler, domains ...string) *host {
	h := &host{
		domains:   make([]string, 0, len(domains)),
		wildcards: make([]string, 0, len(domains)),
		handler:   next,
	}

	for _, domain := range domains {
		if strings.HasPrefix(domain, "*.") {
			h.wildcards = append(h.wildcards, domain[1:]) // 保留 . 符号
		} else {
			h.domains = append(h.domains, domain)
		}
	}

	return h
}

// 查找 hostname 是否与当前的域名匹配。
func (h *host) matched(hostname string) bool {
	index := strings.IndexByte(hostname, ':')
	if index >= 0 {
		hostname = hostname[:index]
	}

	for _, domain := range h.domains {
		if domain == hostname {
			return true
		}
	}

	for _, wildcard := range h.wildcards {
		if strings.HasSuffix(hostname, wildcard) {
			return true
		}
	}

	return false
}

func (h *host) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// r.URL.Hostname() 可能是空值
	if h.matched(r.Host) {
		h.handler.ServeHTTP(w, r)
		return
	}

	http.NotFound(w, r)
}
