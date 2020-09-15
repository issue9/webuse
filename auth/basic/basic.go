// SPDX-License-Identifier: MIT

// Package basic 实现 Basic 校验
//
// https://tools.ietf.org/html/rfc7617
package basic

import (
	"bytes"
	"context"
	"encoding/base64"
	"log"
	"net/http"
	"strings"

	"github.com/issue9/middleware/auth"
)

// AuthFunc 验证登录用户的函数签名
//
// username,password 表示用户登录信息。
// 返回值中，ok 表示是否成功验证。如果成功验证，
// 则 v 用户希望传递给用户的一些额外信息，比如登录用户的权限组什么的。
type AuthFunc func(username, password []byte) (v interface{}, ok bool)

// Basic 验证中间件
type Basic struct {
	auth   AuthFunc
	realm  string
	errlog *log.Logger

	authorization         string
	authenticate          string
	unauthorizationStatus int
}

// New 声明一个 Basic 验证的中间件
//
// proxy 是否为代理，主要是报头的输出内容不同，判断方式完全相同。
// true 会输出 Proxy-Authorization 和 Proxy-Authenticate 报头和 407 状态码，
// 而 false 则是输出 Authorization 和 WWW-Authenticate 报头和 401 状态码；
// log 如果不为 nil，则在运行过程中的错误，将输出到此日志。
func New(auth AuthFunc, realm string, proxy bool, log *log.Logger) *Basic {
	if auth == nil {
		panic("auth 参数不能为空")
	}

	authorization := "Authorization"
	authenticate := "WWW-Authenticate"
	status := http.StatusUnauthorized
	if proxy {
		authorization = "Proxy-Authorization"
		authenticate = "Proxy-Authenticate"
		status = http.StatusProxyAuthRequired
	}

	return &Basic{
		auth:   auth,
		realm:  `Basic realm="` + realm + `"`,
		errlog: log,

		authorization:         authorization,
		authenticate:          authenticate,
		unauthorizationStatus: status,
	}
}

// MiddlewareFunc 将当前中间件应用于 next
func (b *Basic) MiddlewareFunc(next func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return b.Middleware(http.HandlerFunc(next))
}

// Middleware 将当前中间件应用于 next
func (b *Basic) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get(b.authorization)
		index := strings.IndexByte(header, ' ')

		if index <= 0 || index >= len(header) || header[:index] != "Basic" {
			b.unauthorization(w)
			return
		}

		secret, err := base64.StdEncoding.DecodeString(header[index+1:])
		if err != nil {
			if b.errlog != nil {
				b.errlog.Println(err)
			}
			b.unauthorization(w)
			return
		}

		index = bytes.IndexByte(secret, ':')
		if index <= 0 {
			b.unauthorization(w)
			return
		}

		v, ok := b.auth(secret[:index], secret[index+1:])
		if !ok {
			b.unauthorization(w)
			return
		}

		ctx := context.WithValue(r.Context(), auth.ValueKey, v)
		next.ServeHTTP(w, r.WithContext(ctx))
	})

}

func (b *Basic) unauthorization(w http.ResponseWriter) {
	w.Header().Set(b.authenticate, b.realm)
	http.Error(w, http.StatusText(b.unauthorizationStatus), b.unauthorizationStatus)
}
