// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package digest 实现 digest 验证
//
// https://tools.ietf.org/html/rfc7616
package digest

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
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

type digest struct {
	next   http.Handler
	realm  string
	errlog *log.Logger
	nonces *nonces

	authorization         string
	authenticate          string
	unauthorizationStatus int
}

// New 声明一个摘要验证的中间件。
func New(next http.Handler, realm string, proxy bool, errlog *log.Logger) http.Handler {
	authorization := "Authorization"
	authenticate := "WWW-Authenticate"
	status := http.StatusUnauthorized
	if proxy {
		authorization = "Proxy-Authorization"
		authenticate = "Proxy-Authenticate"
		status = http.StatusProxyAuthRequired
	}

	nonces, err := newNonces(24*time.Hour, time.Minute)
	if err != nil {
		panic(err)
	}

	return &digest{
		next:   next,
		realm:  realm,
		errlog: errlog,
		nonces: nonces,

		authorization:         authorization,
		authenticate:          authenticate,
		unauthorizationStatus: status,
	}
}

func (d *digest) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	kv, err := parse(r.Header.Get("Authorization"))
	if err != nil {
		d.errlog.Println(err)
		d.unauthorization(w)
		return
	}

	// TODO
}

func (d *digest) unauthorization(w http.ResponseWriter) {
	var digest strings.Builder

	digest.WriteString(`Digest realm="`)
	digest.WriteString(d.realm)
	digest.WriteString(`",`)

	digest.WriteString(`qop="auth,auth-int",`)

	nonce := d.nonces.newNonce()
	digest.WriteString(`nonce="`)
	digest.WriteString(nonce.key)
	digest.WriteByte('"')

	w.Header().Set(d.authenticate, digest.String())
	http.Error(w, http.StatusText(d.unauthorizationStatus), d.unauthorizationStatus)
}

func parse(digest string) (map[string]string, error) {
	pairs := strings.Split(digest, ",")
	ret := make(map[string]string, len(pairs))

	for _, v := range pairs {
		index := strings.IndexByte(v, '=')
		if index <= 0 {
			return nil, fmt.Errorf("格式错误：%s", digest)
		}

		// TODO 越界检测
		ret[v[:index]] = v[index+2 : len(v)-1]
	}

	return ret, nil
}
