// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package temporary

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/issue9/assert/v4"
	"github.com/issue9/mux/v9/header"
	"github.com/issue9/web"
	"github.com/issue9/web/openapi"
	"github.com/issue9/web/server/servertest"

	"github.com/issue9/webuse/v7/internal/testserver"
	"github.com/issue9/webuse/v7/middlewares/auth"
)

var _ auth.Auth[string] = &Temporary[string]{}

func TestTemporary_header(t *testing.T) {
	a := assert.New(t, false)
	s := testserver.New(a)

	temp := New[string](s, 3*time.Second, false, "", web.ProblemForbidden, web.ProblemBadRequest)
	a.NotNil(temp)
	s.Routers()

	r := s.Routers().New("default", nil)
	r.Post("/login", func(ctx *web.Context) web.Responser {
		return temp.New(ctx, "5", http.StatusCreated)
	})

	r.Get("/info", func(ctx *web.Context) web.Responser {
		if info, ok := temp.GetInfo(ctx); ok {
			return web.OK(info) // info == /login 中传递的值 "5"
		}
		panic("永远不可能达到此处")
	}, temp)

	r.Delete("/login", func(ctx *web.Context) web.Responser {
		if err := temp.Logout(ctx); err != nil {
			return ctx.Error(err, "")
		}
		return web.NoContent()
	}, temp)

	defer servertest.Run(a, s)()
	defer s.Close(0)

	// 未登录
	servertest.Get(a, "http://localhost:8080/info").
		Do(nil).
		Status(http.StatusForbidden)

	servertest.Post(a, "http://localhost:8080/login", nil).
		Do(nil).
		Status(http.StatusCreated).
		BodyFunc(func(a *assert.Assertion, body []byte) {
			resp := &Response{}
			a.NotError(json.Unmarshal(body, resp)).
				NotEmpty(resp.Token).
				Equal(3, resp.Expire)

			// 正常访问
			servertest.Get(a, "http://localhost:8080/info").
				Header(header.Authorization, auth.BearerToken(resp.Token)).
				Do(nil).
				Status(http.StatusOK).
				StringBody(`"5"`)

			// 可多次访问
			servertest.Get(a, "http://localhost:8080/info").
				Header(header.Authorization, auth.BearerToken(resp.Token)).
				Do(nil).
				Status(http.StatusOK).
				StringBody(`"5"`)

			// 删除令牌
			servertest.Delete(a, "http://localhost:8080/login").
				Header(header.Authorization, auth.BearerToken(resp.Token)).
				Do(nil).
				Status(http.StatusNoContent)

			// 不可再次使用
			servertest.Get(a, "http://localhost:8080/info").
				Header(header.Authorization, auth.BearerToken(resp.Token)).
				Do(nil).
				Status(http.StatusForbidden)
		})
}

func TestTemporary_query(t *testing.T) {
	a := assert.New(t, false)
	s := testserver.New(a)

	temp := New[string](s, time.Second, true, "token", web.ProblemForbidden, web.ProblemBadRequest)
	a.NotNil(temp)
	s.Routers()

	r := s.Routers().New("default", nil)
	r.Post("/login", func(ctx *web.Context) web.Responser {
		return temp.New(ctx, "5", http.StatusCreated)
	})

	r.Get("/info", func(ctx *web.Context) web.Responser {
		if info, ok := temp.GetInfo(ctx); ok {
			return web.OK(info) // info == /login 中传递的值 "5"
		}
		panic("永远不可能达到此处")
	}, temp)

	defer servertest.Run(a, s)()
	defer s.Close(0)

	// 未登录
	servertest.Get(a, "http://localhost:8080/info").
		Do(nil).
		Status(http.StatusForbidden)

	servertest.Post(a, "http://localhost:8080/login", nil).
		Do(nil).
		Status(http.StatusCreated).
		BodyFunc(func(a *assert.Assertion, body []byte) {
			resp := &Response{}
			a.NotError(json.Unmarshal(body, resp)).
				NotEmpty(resp.Token).
				Equal(1, resp.Expire)

			// 非查询参数方式，无法通过验证。
			servertest.Get(a, "http://localhost:8080/info").
				Header(header.Authorization, auth.BearerToken(resp.Token)).
				Do(nil).
				Status(http.StatusForbidden)

			// 正常访问
			servertest.Get(a, "http://localhost:8080/info?token="+resp.Token).
				Do(nil).
				Status(http.StatusOK).
				StringBody(`"5"`)

			// 再次访问，令牌失效
			servertest.Get(a, "http://localhost:8080/info?token="+resp.Token).
				Do(nil).
				Status(http.StatusForbidden)
		})
}

func TestSecurityScheme(t *testing.T) {
	a := assert.New(t, false)

	s := SecurityScheme("id", web.Phrase("ss"), "")
	a.Equal(s.Type, openapi.SecuritySchemeTypeHTTP)

	s = SecurityScheme("id", web.Phrase("ss"), "query")
	a.Equal(s.Type, openapi.SecuritySchemeTypeAPIKey).
		Equal(s.In, openapi.InQuery)
}
