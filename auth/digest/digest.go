// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package digest 实现 digest 验证。
// NOTE: 这是个未完成的功能，请勿使用
//
// https://tools.ietf.org/html/rfc7616
//
// TODO: Authorization-Info 等输出的处理
package digest

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/issue9/middleware/auth"
)

func encodeMD5(str string) string {
	m := md5.New()
	m.Write([]byte(str))
	return hex.EncodeToString(m.Sum(nil))
}

// AuthFunc 查找到指定名称的用户数据。
//
// username 表示用户名称；
// v 表示在验证成功的情况下，希望附加到 Request.Context 中的数据；
// password 表示该用户对应的密码；
// found 表示是否找到数据；
type AuthFunc func(username string) (v interface{}, password string, found bool)

type digest struct {
	next   http.Handler
	auth   AuthFunc
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
func New(next http.Handler, auth AuthFunc, realm string, proxy bool, errlog *log.Logger) http.Handler {
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
	if err = nonce.setCount(count); err != nil {
		return nil, err
	}

	obj, pass, ok := d.auth(ret["username"])
	if !ok {
		return nil, errors.New("不存在该用户")
	}
	ha1 := encodeMD5(strings.Join([]string{ret["username"], d.realm, pass}, ":"))
	var ha2 string

	switch ret["qop"] {
	case "auth", "":
		ha2 = encodeMD5(r.Method + ":" + ret["uri"])
	case "auth-int":
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		ha2 = encodeMD5(r.Method + ":" + ret["uri"] + ":" + string(body))
	}

	var resp string
	switch ret["qop"] {
	case "auth", "auth-int":
		resp = encodeMD5(strings.Join([]string{ha1, nonce.key, ret["nc"], ret["cnonce"], ret["qop"], ha2}, ":"))
	default:
		resp = encodeMD5(ha1 + ":" + nonce.key + ":" + ha2)
	}

	if resp == ret["response"] {
		return obj, nil
	}
	return nil, errors.New("验证无法通过")
}

func parseAuthorization(header string) (map[string]string, error) {
	if !strings.HasPrefix(header, "Digest ") {
		return nil, errors.New("无效的起始字符串")
	}

	header = header[7:]
	if len(header) == 0 {
		return nil, errors.New("Authorization 报头内容为空")
	}

	pairs := strings.Split(header, ",")
	ret := make(map[string]string, len(pairs))

	for _, v := range pairs {
		index := strings.IndexByte(v, '=')
		if index <= 0 || index >= len(v)-1 {
			return nil, fmt.Errorf("格式错误：%s", header)
		}

		k := v[:index]
		if _, found := ret[k]; found {
			return nil, fmt.Errorf("存在相同的键名：%s", ret[k])
		}
		ret[k] = v[index+2 : len(v)-1]
	}

	return ret, nil
}
