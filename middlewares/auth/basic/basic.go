// SPDX-FileCopyrightText: 2015-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package basic 实现 [Basic] 校验
//
// [Basic]: https://tools.ietf.org/html/rfc7617
package basic

import (
	"bytes"
	"encoding/base64"
	"net/http"

	"github.com/issue9/mux/v9/header"
	"github.com/issue9/web"
	"github.com/issue9/web/openapi"

	"github.com/issue9/webuse/v7/internal/mauth"
	"github.com/issue9/webuse/v7/middlewares/auth"
)

// AuthFunc 验证登录用户的函数签名
//
// username,password 表示用户登录信息。
// 返回值中，ok 表示是否成功验证。如果成功验证，
// 则 v 为希望传递给用户的一些额外信息，比如登录用户的权限组什么的。
type AuthFunc[T any] func(username, password []byte) (v T, ok bool)

// basic 验证中间件
type basic[T any] struct {
	srv web.Server

	auth  AuthFunc[T]
	realm string

	authorization string
	authenticate  string
	problemID     string
}

// New 声明一个 [Basic 验证]的中间件
//
// proxy 是否为代理，主要是报头的输出内容不同，判断方式完全相同。
// true 会输出 Proxy-Authorization 和 Proxy-Authenticate 报头和 407 状态码，
// 而 false 则是输出 Authorization 和 WWW-Authenticate 报头和 401 状态码；
//
// T 表示验证成功之后，向用户传递的一些额外信息。之后可通过 [GetValue] 获取。
//
// [Basic 验证]: https://datatracker.ietf.org/doc/html/rfc7617
func New[T any](srv web.Server, af AuthFunc[T], realm string, proxy bool) auth.Auth[T] {
	if af == nil {
		panic("auth 参数不能为空")
	}

	authorization := header.Authorization
	authenticate := header.WWWAuthenticate
	problemID := web.ProblemUnauthorized
	if proxy {
		authorization = header.ProxyAuthorization
		authenticate = header.ProxyAuthenticate
		problemID = web.ProblemProxyAuthRequired
	}

	return &basic[T]{
		srv: srv,

		auth:  af,
		realm: auth.BasicToken(`realm="` + realm + `"`),

		authorization: authorization,
		authenticate:  authenticate,
		problemID:     problemID,
	}
}

func (b *basic[T]) Middleware(next web.HandlerFunc, method, _, _ string) web.HandlerFunc {
	if method == http.MethodOptions {
		return next
	}

	return func(ctx *web.Context) web.Responser {
		h := auth.GetBasicToken(ctx, b.authorization)

		secret, err := base64.StdEncoding.DecodeString(h)
		if err != nil {
			ctx.Header().Set(b.authenticate, b.realm)
			return ctx.Error(err, b.problemID)
		}

		pp, ss, ok := bytes.Cut(secret, []byte{':'})
		if !ok {
			return b.unauthorization(ctx)
		}
		v, ok := b.auth(pp, ss)
		if !ok {
			return b.unauthorization(ctx)
		}

		mauth.Set(ctx, v)
		return next(ctx)
	}
}

func (b *basic[T]) Logout(*web.Context) error { return nil }

func (b *basic[T]) unauthorization(ctx *web.Context) web.Responser {
	ctx.Header().Set(b.authenticate, b.realm)
	return ctx.Problem(b.problemID)
}

func (b *basic[T]) GetInfo(ctx *web.Context) (T, bool) { return mauth.Get[T](ctx) }

// SecurityScheme 声明支持 openapi 的 [openapi.SecurityScheme] 对象
func SecurityScheme(id string, desc web.LocaleStringer) *openapi.SecurityScheme {
	return &openapi.SecurityScheme{
		ID:          id,
		Type:        openapi.SecuritySchemeTypeHTTP,
		Description: desc,
		Scheme:      auth.Basic[:len(auth.Basic)-1],
	}
}
