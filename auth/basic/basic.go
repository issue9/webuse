// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package basic 实现 Basic 校验
package basic

import (
	"bytes"
	"context"
	"encoding/base64"
	"log"
	"net/http"
	"strings"
)

type keyType int

// ValueKey 保存于 context 中的值的名称
const ValueKey keyType = 0

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
// 而 false 则是输出 Authorization 和 WWW-Authenticate 报头和 401 状态码。
func New(next http.Handler, auth AuthFunc, realm string, proxy bool, log *log.Logger) http.Handler {
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
	auth := r.Header.Get(b.authorization)
	index := strings.IndexByte(auth, ' ')

	if index <= 0 || index >= len(auth) || auth[:index] != "Basic" {
		b.unauthorization(w)
		return
	}

	secret, err := base64.StdEncoding.DecodeString(auth[index+1:])
	if err != nil {
		b.errlog.Println(err)
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

	ctx := context.WithValue(r.Context(), ValueKey, v)
	r = r.WithContext(ctx)

	b.next.ServeHTTP(w, r)
}

func (b *basic) unauthorization(w http.ResponseWriter) {
	w.Header().Set(b.authenticate, b.realm)
	http.Error(w, http.StatusText(b.unauthorizationStatus), b.unauthorizationStatus)
}
