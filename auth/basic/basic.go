// SPDX-License-Identifier: MIT

// Package basic 实现 [Basic] 校验
//
// [Basic]: https://tools.ietf.org/html/rfc7617
package basic

import (
	"bytes"
	"encoding/base64"
	"strings"

	"github.com/issue9/web"
)

type keyType int

const valueKey keyType = 1

const prefix = "basic "

const prefixLen = 6 // len(prefix)

// AuthFunc 验证登录用户的函数签名
//
// username,password 表示用户登录信息。
// 返回值中，ok 表示是否成功验证。如果成功验证，
// 则 v 为希望传递给用户的一些额外信息，比如登录用户的权限组什么的。
type AuthFunc[T any] func(username, password []byte) (v T, ok bool)

// basic 验证中间件
type basic[T any] struct {
	srv *web.Server

	auth  AuthFunc[T]
	realm string

	authorization string
	authenticate  string
	problemID     string
}

// New 声明一个 Basic 验证的中间件
//
// proxy 是否为代理，主要是报头的输出内容不同，判断方式完全相同。
// true 会输出 Proxy-Authorization 和 Proxy-Authenticate 报头和 407 状态码，
// 而 false 则是输出 Authorization 和 WWW-Authenticate 报头和 401 状态码；
//
// T 表示验证成功之后，向用户传递的一些额外信息。之后可通过 [Basic.GetValue] 获取。
func New[T any](srv *web.Server, auth AuthFunc[T], realm string, proxy bool) web.Middleware {
	if auth == nil {
		panic("auth 参数不能为空")
	}

	authorization := "Authorization"
	authenticate := "WWW-Authenticate"
	problemID := web.ProblemUnauthorized
	if proxy {
		authorization = "Proxy-Authorization"
		authenticate = "Proxy-Authenticate"
		problemID = web.ProblemProxyAuthRequired
	}

	return &basic[T]{
		srv: srv,

		auth:  auth,
		realm: `Basic realm="` + realm + `"`,

		authorization: authorization,
		authenticate:  authenticate,
		problemID:     problemID,
	}
}

func (b *basic[T]) Middleware(next web.HandlerFunc) web.HandlerFunc {
	return func(ctx *web.Context) web.Responser {
		h := ctx.Request().Header.Get(b.authorization)
		if len(h) > prefixLen && strings.ToLower(h[:prefixLen]) == prefix {
			h = h[prefixLen:]
		}

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
		ctx.SetVar(valueKey, v)

		return next(ctx)
	}
}

func (b *basic[T]) unauthorization(ctx *web.Context) web.Responser {
	ctx.Header().Set(b.authenticate, b.realm)
	return ctx.Problem(b.problemID)
}

// GetValue 获取当前对话关联的登录信息
func GetValue[T any](ctx *web.Context) (T, bool) {
	if v, found := ctx.GetVar(valueKey); found {
		return v.(T), true
	}
	var vv T
	return vv, false
}
