// SPDX-FileCopyrightText: 202402025 caixw
//
// SPDX-License-Identifier: MIT

// Package temporary 用于创建一个一次性的令牌
package temporary

import (
	"errors"
	"net/http"
	"time"

	"github.com/issue9/cache"
	"github.com/issue9/mux/v9/header"
	"github.com/issue9/web"
	"github.com/issue9/web/openapi"

	"github.com/issue9/webuse/v7/internal/mauth"
	"github.com/issue9/webuse/v7/middlewares/auth"
)

type tokenType int

const tokenContext tokenType = 0

type Response struct {
	XMLName struct{} `json:"-" cbor:"-" xml:"token" yaml:"-"`
	Token   string   `json:"token" xml:"token" cbor:"token" yaml:"token" comment:"access token"`                  // 访问令牌
	Expire  int      `json:"expire" xml:"expire,attr" cbor:"expire" yaml:"expire" comment:"access token expired"` // 访问令牌的有效时长，单位为秒
}

type Temporary[T any] struct {
	cache                 web.Cache
	ttl                   time.Duration
	expire                int
	once                  bool
	query                 string
	unauthProblemID       string
	invalidTokenProblemID string
}

// New 创建 [Temporary] 对象
//
// ttl 表示令牌的过期时间。
// once 是否为一次性令牌，如果为 true，在验证成功之后，该令牌将自动失效；
// query 如果不为空，那么将由查询参数传递验证，否则表示 Bearer 类型的令牌传递；
// unauthProblemID 验证不通过时的错误代码；
// invalidTokenProblemID 令牌无效时返回的错误代码；
func New[T any](s web.Server, ttl time.Duration, once bool, query string, unauthProblemID, invalidTokenProblemID string) *Temporary[T] {
	return &Temporary[T]{
		cache:                 web.NewCache(s.UniqueID(), s.Cache()),
		ttl:                   ttl,
		expire:                int(ttl.Seconds()),
		once:                  once,
		query:                 query,
		unauthProblemID:       unauthProblemID,
		invalidTokenProblemID: invalidTokenProblemID,
	}
}

// New 创建令牌
//
// v 为令牌关联的数据，之后通过验证接口可以访问该数据；
func (t *Temporary[T]) New(ctx *web.Context, v T, status int) web.Responser {
	token := ctx.Server().UniqueID()
	if err := t.cache.Set(token, v, t.ttl); err != nil {
		return ctx.Error(err, "")
	}

	return web.Response(status, &Response{Token: token, Expire: t.expire})
}

func (t *Temporary[T]) Middleware(next web.HandlerFunc, method, _, _ string) web.HandlerFunc {
	if method == http.MethodOptions {
		return next
	}

	return func(ctx *web.Context) web.Responser {
		var token string
		if t.query != "" {
			q, err := ctx.Queries(true)
			if err != nil {
				return ctx.Problem(t.invalidTokenProblemID)
			}
			token = q.String(t.query, "")
		} else {
			token = auth.GetBearerToken(ctx, header.Authorization)
		}
		if token == "" {
			return ctx.Problem(t.unauthProblemID)
		}

		var v T
		err := t.cache.Get(token, &v)
		switch {
		case errors.Is(err, cache.ErrCacheMiss()):
			return ctx.Problem(t.unauthProblemID)
		case err != nil:
			return ctx.Error(err, t.invalidTokenProblemID)
		default:
			mauth.Set(ctx, v)
			ctx.SetVar(tokenContext, token)

			if t.once {
				if err := t.cache.Delete(token); err != nil {
					ctx.Server().Logs().ERROR().Error(err) // 只记录错误，不反馈给客户端。
				}
			}

			return next(ctx)
		}
	}
}

func (t *Temporary[T]) Logout(ctx *web.Context) error {
	if key, found := ctx.GetVar(tokenContext); found {
		return t.cache.Delete(key.(string))
	}
	return nil
}

// QueryName 查询参数的名称
func (t *Temporary[T]) QueryName() string { return t.query }

func (t *Temporary[T]) GetInfo(ctx *web.Context) (T, bool) { return mauth.Get[T](ctx) }

func (t *Temporary[T]) SecurityScheme(id string, desc web.LocaleStringer) *openapi.SecurityScheme {
	return SecurityScheme(id, desc, t.QueryName())
}

// SecurityScheme 声明支持 openapi 的 [openapi.SecurityScheme] 对象
func SecurityScheme(id string, desc web.LocaleStringer, query string) *openapi.SecurityScheme {
	if query != "" {
		return &openapi.SecurityScheme{
			ID:          id,
			Type:        openapi.SecuritySchemeTypeAPIKey,
			Description: desc,
			In:          openapi.InQuery,
			Name:        query,
		}
	}

	return &openapi.SecurityScheme{
		ID:          id,
		Type:        openapi.SecuritySchemeTypeHTTP,
		Description: desc,
		Scheme:      auth.Bearer[:len(auth.Bearer)-1],
	}
}
