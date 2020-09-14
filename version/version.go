// SPDX-License-Identifier: MIT

// Package version 提供一个限定版本号的中间件
package version

import (
	"net/http"
	"strings"
)

const versionString = "version="

// Version 限定版本号的中间件
//
// 从请求报头的 Accept 中解析相应的版本号，不区分大小写。
//
// 当版本号不匹配时，返回 404 错误信息。
//
// 若要将版本号放在路径中，可以直接使用 https://github.com/issue9/mux.Prefix 对象
type Version struct {
	Version string
	Strict  bool // 在没有指定版本号时的处理方式，是否统一采用拒绝访问。
}

// Middleware 将当前中间件应用于 next
func (v *Version) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ver := findVersionNumber(r.Header.Get("Accept"))

		if len(ver) == 0 {
			if v.Strict { // strict 模式下
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			} else {
				next.ServeHTTP(w, r)
			}
			return
		}

		if ver != v.Version {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Middleware 将当前中间件应用于 f
func (v *Version) MiddlewareFunc(f func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return v.Middleware(http.HandlerFunc(f))
}

// 从 accept 中找到版本号，或是没有找到时，返回第二个参数 false。
func findVersionNumber(accept string) string {
	strs := strings.Split(accept, ";")
	for _, str := range strs {
		str = strings.ToLower(strings.TrimSpace(str))
		if index := strings.Index(str, versionString); index >= 0 {
			return str[index+len(versionString):]
		}
	}

	return ""
}
