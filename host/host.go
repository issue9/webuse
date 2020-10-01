// SPDX-License-Identifier: MIT

// Package host 提供了限定访问域名的中间件
package host

import (
	"net/http"
	"strings"

	"github.com/issue9/sliceutil"
)

// Host 限定域名的中间件
type Host struct {
	omitempty bool     // 在限定条件为空时，对所有访问一律放行
	domains   []string // 域名列表
	wildcards []string // 泛域名列表，只保存 * 之后的部分内容
}

// New 声明一个限定域名的中间件
//
// 若请求的域名不允许，会返回 404 错误。
// 若 domain 为空，则任何请求都将返回 404。
//
// 仅会将域名与 domains 进行比较，端口与协议都将不参写比较。
func New(omitempty bool, domain ...string) *Host {
	h := &Host{
		omitempty: omitempty,
		domains:   make([]string, 0, len(domain)),
		wildcards: make([]string, 0, len(domain)),
	}

	h.Add(domain...)

	return h
}

// Add 添加新的域名
//
// domain 可以是泛域名，比如 *.example.com，但不能是 s1.*.example.com。
//
// NOTE: 重复的值不会重复添加。
func (h *Host) Add(domain ...string) {
	for _, d := range domain {
		switch {
		case strings.HasPrefix(d, "*."):
			d = d[1:] // 保留 . 符号
			if sliceutil.Count(h.wildcards, func(i int) bool { return d == h.wildcards[i] }) <= 0 {
				h.wildcards = append(h.wildcards, d)
			}
		default:
			if sliceutil.Count(h.domains, func(i int) bool { return d == h.domains[i] }) <= 0 {
				h.domains = append(h.domains, d)
			}
		}
	}
}

// Delete 删除域名
//
// NOTE: 如果不存在，则不作任何改变。
func (h *Host) Delete(domain string) {
	switch {
	case strings.HasPrefix(domain, "*."):
		size := sliceutil.Delete(h.wildcards, func(i int) bool { return h.wildcards[i] == domain[1:] })
		h.wildcards = h.wildcards[:size]
	default:
		size := sliceutil.Delete(h.domains, func(i int) bool { return h.domains[i] == domain })
		h.domains = h.domains[:size]
	}
}

// 查找 hostname 是否与当前的域名匹配
func (h *Host) matched(hostname string) bool {
	if len(h.domains) == 0 && len(h.wildcards) == 0 && h.omitempty {
		return true
	}

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

// Middleware 将当前中间件应用于 next
func (h *Host) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// r.URL.Hostname() 可能是空值
		if h.matched(r.Host) {
			next.ServeHTTP(w, r)
			return
		}

		http.NotFound(w, r)
	})
}

// MiddlewareFunc 将当前中间件应用于 next
func (h *Host) MiddlewareFunc(next func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return h.Middleware(http.HandlerFunc(next))
}
