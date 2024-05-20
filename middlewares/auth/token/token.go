// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

// Package token 传统令牌的验证方式
package token

import (
	"errors"
	"time"

	"github.com/issue9/mux/v8/header"
	"github.com/issue9/rands/v3"
	"github.com/issue9/web"

	"github.com/issue9/webuse/v7/internal/mauth"
	"github.com/issue9/webuse/v7/middlewares/auth"
)

type tokenType int

const tokenContext tokenType = 0

// Token 传统的令牌管理
//
// 每次产生两个令牌：
// 访问令牌可用于普通的中间件验证；
// 刷新令牌也可以用于验证，但是是一次性的，验证完，刷新令牌和关联的访问令牌都将失效果，
// 所以刷新令牌一般只能用于申请下一次的令牌；
//
// T 为每次登录之后需要与令牌关联的数据，在登录失效之前，
// [Store] 将一直保留该数据，不用再次访问数据系统。
type Token[T UserData] struct {
	s     web.Server
	rands *rands.Rands[byte]
	store Store[T]
	br    BuildResponseFunc

	accessExp, refreshExp       time.Duration
	accessExpInt, refreshExpInt int
	invalidTokenProblemID       string
}

// New 声明 [Token] 对象
//
// accessExp，refreshExp 表示访问令牌和刷新令牌的有效时长，refreshExp 必须大于 accessExp；
// invalidTokenProblemID 令牌无效时返回的错误代码。比如将访问令牌当刷新令牌使用等；
// br 用于生成向客户端反馈令牌信息的结构体方法，默认为 [DefaultBuildResponse]；
func New[T UserData](
	s web.Server,
	store Store[T],
	accessExp, refreshExp time.Duration,
	invalidTokenProblemID string,
	br BuildResponseFunc,
) *Token[T] {
	if accessExp >= refreshExp {
		panic("参数 accessExp 必须小于 refreshExp")
	}
	if br == nil {
		br = DefaultBuildResponse
	}

	r := rands.New(nil, 100, 15, 16, rands.AlphaNumber())
	s.Services().Add(web.Phrase("gen token id"), r)

	return &Token[T]{
		s:     s,
		rands: r,
		store: store,
		br:    br,

		accessExp:             accessExp,
		refreshExp:            refreshExp,
		accessExpInt:          int(accessExp.Seconds()),
		refreshExpInt:         int(refreshExp.Seconds()),
		invalidTokenProblemID: invalidTokenProblemID,
	}
}

func (t *Token[T]) Middleware(next web.HandlerFunc, _, _ string) web.HandlerFunc {
	return func(ctx *web.Context) web.Responser {
		token := auth.GetToken(ctx, auth.Bearer, header.Authorization)
		if token == "" {
			return ctx.Problem(web.ProblemUnauthorized)
		}

		switch v, found, err := t.store.Get(token); {
		case err != nil:
			return ctx.Error(err, t.invalidTokenProblemID)
		case !found:
			return ctx.Problem(web.ProblemUnauthorized)
		default:
			mauth.Set(ctx, v)
			ctx.SetVar(tokenContext, token) // 先保存到上下文环境中，再从存储系统中删除令牌信息。

			if v.Access != "" { // 刷新令牌
				// 删除令牌出错，不防碍其它功能的继承执行，所以只记录日志，不退出。
				if err := errors.Join(t.store.DeleteToken(v.Access), t.store.DeleteToken(token)); err != nil {
					t.s.Logs().ERROR().Error(err)
				}
			}

			return next(ctx)
		}
	}
}

func (t *Token[T]) Logout(ctx *web.Context) error {
	if key, found := ctx.GetVar(tokenContext); found {
		return t.store.DeleteToken(key.(string))
	}
	return nil
}

func (t *Token[T]) GetInfo(ctx *web.Context) (T, bool) {
	if v, found := mauth.Get[Item[T]](ctx); found {
		return v.UserData, true
	}
	var zero T
	return zero, false
}

// New 根据给定的参数 v 创建新的令牌
//
// v 为新令牌需要关联的值；
// status 为输出的状态码；
// headers 报头列表，第一个元素为报头，第二个元素为对应的值，依次类推；
func (t *Token[T]) New(ctx *web.Context, v T, status int, headers ...string) web.Responser {
	access := t.s.UniqueID() + t.rands.String()
	accessItem := Item[T]{UserData: v}
	refresh := t.s.UniqueID() + t.rands.String()
	refreshItem := Item[T]{UserData: v, Access: access}

	if err := t.store.Save(access, accessItem, t.accessExp); err != nil {
		return ctx.Error(err, "")
	}
	if err := t.store.Save(refresh, refreshItem, t.refreshExp); err != nil {
		return ctx.Error(err, "")
	}

	return web.Response(status, t.br(access, refresh, t.accessExpInt, t.refreshExpInt), headers...)
}

// Refresh 刷新令牌
func (t *Token[T]) Refresh(ctx *web.Context, status int, headers ...string) web.Responser {
	v, found := mauth.Get[Item[T]](ctx)
	if !found {
		panic("通过了令牌验证但是在 Context 找不到相关信息")
	}

	if v.Access == "" { // 不是刷新令牌
		return ctx.Problem(t.invalidTokenProblemID)
	}
	return t.New(ctx, v.UserData, status, headers...)
}

// Delete 根据指定的用户数据
func (t *Token[T]) Delete(u T) error { return t.store.DeleteUID(u.GetUID()) }
