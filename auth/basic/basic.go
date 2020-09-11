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

type basic struct {
	next   http.Handler
	auth   AuthFunc
	realm  string
	errlog *log.Logger

	authorization         string
	authenticate          string
	unauthorizationStatus int
}

// New 声明一个 Basic 验证的中间件
//
// next 表示验证通过之后，需要执行的 handler；
// proxy 是否为代码，主要是报头的输出内容不同，判断方式完全相同。
// true 会输出 Proxy-Authorization 和 Proxy-Authenticate 报头和 407 状态码，
// 而 false 则是输出 Authorization 和 WWW-Authenticate 报头和 401 状态码；
// log 如果不为 nil，则在运行过程中的错误，将输出到此日志。
func New(next http.Handler, auth AuthFunc, realm string, proxy bool, log *log.Logger) http.Handler {
	if next == nil {
		panic("next 参数不能为空")
	}

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

	return &basic{
		next:   next,
		auth:   auth,
		realm:  `Basic realm="` + realm + `"`,
		errlog: log,

		authorization:         authorization,
		authenticate:          authenticate,
		unauthorizationStatus: status,
	}
}

func (b *basic) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	b.next.ServeHTTP(w, r.WithContext(ctx))
}

func (b *basic) unauthorization(w http.ResponseWriter) {
	w.Header().Set(b.authenticate, b.realm)
	http.Error(w, http.StatusText(b.unauthorizationStatus), b.unauthorizationStatus)
}
