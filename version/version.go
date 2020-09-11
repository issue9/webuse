// SPDX-License-Identifier: MIT

// Package version 提供一个限定版本号的中间件
package version

import (
	"net/http"
	"strings"
)

const versionString = "version="

type version struct {
	handler http.Handler
	version string
	strict  bool
}

// New 构建一个限定版本号的中间件
//
// 从请求报头的 Accept 中解析相应的版本号，不区分大小写。
//
// 当版本号不匹配时，返回 404 错误信息。
//
// v 只有与此匹配的版本号，才能运行 h；
// strict 在没有指定版本号时的处理方式，为 false 时，请求头无版本号
// 表示可以匹配；为 true 时，请求头无版本号表示不匹配。
//
// 若要将版本号放在路径中，可以直接使用 https://github.com/issue9/mux.Prefix 对象
func New(next http.Handler, v string, strict bool) http.Handler {
	return &version{
		handler: next,
		version: v,
		strict:  strict,
	}
}

func (v *version) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ver := findVersionNumber(r.Header.Get("Accept"))

	if len(ver) == 0 {
		if v.strict { // strict 模式下
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		} else {
			v.handler.ServeHTTP(w, r)
		}

		return
	}

	if ver != v.version {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	v.handler.ServeHTTP(w, r)
}

// 从 accept 中找到版本号，或是没有找到时，返回第二个参数 false。
func findVersionNumber(accept string) string {
	strs := strings.Split(accept, ";")
	for _, str := range strs {
		str = strings.ToLower(strings.TrimSpace(str))
		index := strings.Index(str, versionString)

		if index >= 0 {
			return str[index+len(versionString):]
		}
	}

	return ""
}
