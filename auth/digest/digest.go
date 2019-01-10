// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package digest 实现 digest 验证
//
// https://tools.ietf.org/html/rfc7616
//
// NOTE: 实验中，未作任何测试
package digest

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/issue9/utils"

	"github.com/issue9/middleware/auth"
)

// Auther 验证用户信息的接口
type Auther interface {
	// 根据用户名，找到其对应的密码
	Password(username string) string

	// 根据用户名，获取其相关的信息，方便附加到 request.Context，传递给其它中间件。
	Object(username string) interface{}
}

type digest struct {
	next   http.Handler
	auth   Auther
	realm  string
	errlog *log.Logger
	nonces *nonces

	authorization         string
	authenticate          string
	unauthorizationStatus int
}

// New 声明一个摘要验证的中间件。
//
// next 表示验证通过之后，需要执行的 handler；
// proxy 是否为代码，主要是报头的输出内容不同，判断方式完全相同。
// true 会输出 Proxy-Authorization 和 Proxy-Authenticate 报头和 407 状态码，
// 而 false 则是输出 Authorization 和 WWW-Authenticate 报头和 401 状态码；
// log 如果不为 nil，则在运行过程中的错误，将输出到此日志。
func New(next http.Handler, auth Auther, realm string, proxy bool, errlog *log.Logger) http.Handler {
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
		auth:   auth,
		realm:  realm,
		errlog: errlog,
		nonces: nonces,

		authorization:         authorization,
		authenticate:          authenticate,
		unauthorizationStatus: status,
	}
}

func (d *digest) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v, err := d.parse(r)
	if err != nil {
		if d.errlog != nil {
			d.errlog.Println(err)
		}
		d.unauthorization(w)
		return
	}

	ctx := context.WithValue(r.Context(), auth.ValueKey, v)
	r = r.WithContext(ctx)
	d.next.ServeHTTP(w, r)
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

func (d *digest) parse(r *http.Request) (interface{}, error) {
	ret, err := parseAuthorization(r.Header.Get("Authorization"))

	// 基本检测
	nonce := d.nonces.get(ret["nonce"])
	if nonce == nil {
		return nil, errors.New("nonce 不存在")
	}

	if ret["realm"] != d.realm {
		return nil, errors.New("realm 不正确")
	}

	count, err := strconv.Atoi(ret["nc"])
	if err != nil {
		return nil, err
	}
	if nonce.count >= count {
		return nil, errors.New("计数器不准确")
	}

	pass := d.auth.Password(ret["username"])
	ha1 := utils.MD5(strings.Join([]string{ret["username"], d.realm, pass}, ":"))
	var ha2 string

	switch ret["qop"] {
	case "auth", "":
		ha2 = utils.MD5(r.Method + ":" + ret["uri"])
	case "auth-int":
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		ha2 = utils.MD5(r.Method + ":" + ret["uri"] + ":" + string(body))
	}

	var resp string
	switch ret["qop"] {
	case "auth", "auth-int":
		resp = utils.MD5(strings.Join([]string{ha1, nonce.key, ret["nc"], ret["cnonce"], ret["qop"], ha2}, ":"))
	default:
		resp = utils.MD5(ha1 + ":" + nonce.key + ":" + ha2)
	}

	if resp == ret["response"] {
		return d.auth.Object(ret["username"]), nil
	}
	return nil, errors.New("验证无法通过")
}

func parseAuthorization(header string) (map[string]string, error) {
	if !strings.HasPrefix(header, "Digest ") {
		return nil, errors.New("无效的起始字符串")
	}

	pairs := strings.Split(header, ",")
	ret := make(map[string]string, len(pairs))

	for _, v := range pairs {
		index := strings.IndexByte(v, '=')
		if index <= 0 || index >= len(v) {
			return nil, fmt.Errorf("格式错误：%s", header)
		}

		ret[v[:index]] = v[index+2 : len(v)-1]
	}

	return ret, nil
}
