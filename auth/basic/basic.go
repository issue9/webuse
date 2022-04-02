// SPDX-License-Identifier: MIT

// Package basic 实现 Basic 校验
//
// https://tools.ietf.org/html/rfc7617
package basic

import (
	"bytes"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/issue9/web"

	"github.com/issue9/middleware/v6/auth"
)

// AuthFunc 验证登录用户的函数签名
//
// username,password 表示用户登录信息。
// 返回值中，ok 表示是否成功验证。如果成功验证，
// 则 v 为用户希望传递给用户的一些额外信息，比如登录用户的权限组什么的。
type AuthFunc func(username, password []byte) (v any, ok bool)

// Basic 验证中间件
type Basic struct {
	srv *web.Server

	auth  AuthFunc
	realm string

	authorization         string
	authenticate          string
	unauthorizationStatus int
}

// New 声明一个 Basic 验证的中间件
//
// proxy 是否为代理，主要是报头的输出内容不同，判断方式完全相同。
// true 会输出 Proxy-Authorization 和 Proxy-Authenticate 报头和 407 状态码，
// 而 false 则是输出 Authorization 和 WWW-Authenticate 报头和 401 状态码；
func New(srv *web.Server, auth AuthFunc, realm string, proxy bool) *Basic {
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
		srv: srv,

		auth:  auth,
		realm: `Basic realm="` + realm + `"`,

		authorization:         authorization,
		authenticate:          authenticate,
		unauthorizationStatus: status,
	}
}

func (b *Basic) Middleware(next web.HandlerFunc) web.HandlerFunc {
	return func(ctx *web.Context) web.Responser {
		header := ctx.Request().Header.Get(b.authorization)

		p, s, ok := strings.Cut(header, " ")
		if !ok || p != "Basic" {
			return b.unauthorization(ctx)
		}

		secret, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			b.srv.Logs().ERROR().Error(err)
			return b.unauthorization(ctx)
		}

		pp, ss, ok := bytes.Cut(secret, []byte{':'})
		if !ok {
			return b.unauthorization(ctx)
		}
		v, ok := b.auth(pp, ss)
		if !ok {
			return b.unauthorization(ctx)
		}
		auth.SetValue(ctx, v)

		return next(ctx)
	}
}

func (b *Basic) unauthorization(ctx *web.Context) web.Responser {
	return ctx.Status(b.unauthorizationStatus, b.authenticate, b.realm)
}
