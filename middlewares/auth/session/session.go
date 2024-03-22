// SPDX-FileCopyrightText: 2015-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package session SESSION 管理
package session

import (
	"errors"
	"net/http"
	"net/url"
	"reflect"
	"time"

	"github.com/issue9/rands/v3"
	"github.com/issue9/unique/v2"
	"github.com/issue9/web"
)

var errSessionIDNotExists = web.NewLocaleError("session id not exists in context")

const contextTypeKey contextType = 0
const idKey contextType = 1

type contextType int

// Session session 管理
type Session[T any] struct {
	rands *unique.Rands
	store Store[T]

	// cookie 的相关设置
	lifetime           int
	name, path, domain string
	secure, httpOnly   bool
}

func ErrSessionKDNotExists() error { return errSessionIDNotExists }

// New 声明 [Session] 中间件
//
// lifetime 为 session 的有效时间，单位为秒；其它参数为 cookie 的相关设置。
func New[T any](s web.Server, store Store[T], lifetime int, name, path, domain string, secure, httpOnly bool) *Session[T] {
	r := unique.NewRands(100, nil, 10, 11, rands.AlphaNumber())
	s.Services().Add(web.Phrase("gen session id"), r)

	return &Session[T]{
		rands: r,
		store: store,

		lifetime: lifetime,
		name:     name,
		path:     path,
		domain:   domain,
		secure:   secure,
		httpOnly: httpOnly,
	}
}

func (s *Session[T]) Middleware(next web.HandlerFunc) web.HandlerFunc {
	return func(ctx *web.Context) web.Responser {
		c, err := ctx.Request().Cookie(s.name)
		if err != nil && !errors.Is(err, http.ErrNoCookie) { // 不退出，给定默认值。
			ctx.Logs().ERROR().Error(err)
		}

		var id string
		if c == nil {
			id = s.rands.String()
			c = &http.Cookie{
				Name:     s.name,
				Path:     s.path,
				Domain:   s.domain,
				Secure:   s.secure,
				HttpOnly: s.httpOnly,
				Value:    url.QueryEscape(id),
			}
		} else {
			id, err = url.QueryUnescape(c.Value)
			if err != nil {
				return ctx.Error(err, web.ProblemInternalServerError)
			}
		}

		c.MaxAge = s.lifetime
		c.Expires = ctx.Begin().Add(time.Second * time.Duration(s.lifetime)) // http 1.0 和 ie8 仅支持此属性
		ctx.SetCookies(c)
		ctx.SetVar(idKey, id)

		v, found, err := s.store.Get(id)
		if err != nil {
			return ctx.Error(err, web.ProblemInternalServerError)
		} else if !found {
			var zero T
			// BUG 多层指针？
			if t := reflect.TypeOf(zero); t.Kind() == reflect.Pointer {
				v = reflect.New(t.Elem()).Interface().(T)
			}

			// 生成 v，需要保存
			if err := s.store.Set(id, v); err != nil {
				return ctx.Error(err, web.ProblemInternalServerError)
			}
		}

		ctx.SetVar(contextTypeKey, v)

		return next(ctx)
	}
}

// Logout 退出登录
func (s *Session[T]) Logout(ctx *web.Context) error {
	id, err := s.GetSessionID(ctx)
	if err == nil {
		err = s.Delete(id)
	}
	return err
}

// Delete 删除 session id
func (s *Session[T]) Delete(sessionid string) error { return s.store.Delete(sessionid) }

func (s *Session[T]) GetSessionID(ctx *web.Context) (string, error) {
	v, found := ctx.GetVar(idKey)
	if !found {
		return "", ErrSessionKDNotExists()
	}
	return v.(string), nil
}

// Save 保存 val
func (s *Session[T]) Save(ctx *web.Context, val T) error {
	SetValue(ctx, val)
	id, err := s.GetSessionID(ctx)
	if err != nil {
		return err
	}
	return s.store.Set(id, val)
}

// GetValue 获取当前对话关联的信息
func GetValue[T any](ctx *web.Context) (val T, err error) {
	v, found := ctx.GetVar(contextTypeKey)
	if found {
		return v.(T), nil
	}

	return v.(T), web.NewLocaleError("not found the context session key")
}

// SetValue 更新 [web.Context] 保存的值
func SetValue[T any](ctx *web.Context, val T) error {
	ctx.SetVar(contextTypeKey, val)
	return nil
}

// DelValue 删除 [web.Context] 中保存的值
func DelValue[T any](ctx *web.Context) error {
	ctx.DelVar(contextTypeKey)
	return nil
}
