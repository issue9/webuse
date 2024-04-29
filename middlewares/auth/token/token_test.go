// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package token

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/issue9/assert/v4"
	"github.com/issue9/mux/v8/header"
	"github.com/issue9/web"
	"github.com/issue9/web/server/servertest"

	"github.com/issue9/webuse/v7/internal/testserver"
	"github.com/issue9/webuse/v7/middlewares/auth"
)

type v struct {
	ID string
}

var _ UserData = v{}

func (v v) GetUID() string { return v.ID }

var _ auth.Auth[v] = &Token[v]{}

func TestToken(t *testing.T) {
	a := assert.New(t, false)
	s := testserver.New(a)

	token := New(s, NewCacheStore[v](s.Cache()), time.Second, 2*time.Second, web.ProblemBadRequest, nil)
	a.NotNil(token)
	s.Routers()

	r := s.Routers().New("default", nil)
	r.Post("/login", func(ctx *web.Context) web.Responser {
		return token.New(ctx, v{ID: "5"}, http.StatusCreated)
	})

	r.Get("/info", token.Middleware(func(ctx *web.Context) web.Responser {
		if info, ok := token.GetInfo(ctx); ok {
			return web.OK(info) // info == /login 中传递的值 v{ID:"5"}
		}
		panic("永远不可能达到此处")
	}))

	r.Post("/refresh", token.Middleware(func(ctx *web.Context) web.Responser {
		return token.Refresh(ctx, http.StatusOK)
	}))

	defer servertest.Run(a, s)()
	defer s.Close(0)

	// 未登录
	servertest.Get(a, "http://localhost:8080/info").
		Do(nil).
		Status(http.StatusUnauthorized)

	servertest.Post(a, "http://localhost:8080/login", nil).
		Do(nil).
		Status(http.StatusCreated).
		BodyFunc(func(a *assert.Assertion, body []byte) {
			resp := &Response{}
			a.NotError(json.Unmarshal(body, resp)).
				NotEmpty(resp.AccessToken).
				NotEmpty(resp.RefreshToken).
				Equal(1, resp.AccessExp).
				Equal(2, resp.RefreshExp)

			// 正常访问
			servertest.Get(a, "http://localhost:8080/info").
				Header(header.Authorization, auth.BuildToken(auth.Bearer, resp.AccessToken)).
				Do(nil).
				Status(http.StatusOK).
				StringBody(`{"ID":"5"}`)

			// 刷新令牌
			servertest.Post(a, "http://localhost:8080/refresh", nil).
				Header(header.Authorization, auth.BuildToken(auth.Bearer, resp.RefreshToken)).
				Do(nil).
				Status(http.StatusOK).
				BodyFunc(func(a *assert.Assertion, body []byte) {
					resp2 := &Response{}
					a.NotError(json.Unmarshal(body, resp2)).
						NotEmpty(resp2.AccessToken).
						NotEmpty(resp2.RefreshToken).
						Equal(1, resp2.AccessExp).
						Equal(2, resp2.RefreshExp)

					// 旧的访问令牌已经不能访问
					servertest.Get(a, "http://localhost:8080/info").
						Header(header.Authorization, auth.BuildToken(auth.Bearer, resp.AccessToken)).
						Do(nil).
						Status(http.StatusUnauthorized) // token 在 /refresh 中已经被删除

					// 旧的刷新令牌已经不能访问
					servertest.Post(a, "http://localhost:8080/refresh", nil).
						Header(header.Authorization, auth.BuildToken(auth.Bearer, resp.RefreshToken)).
						Do(nil).
						Status(http.StatusUnauthorized)

					// 采用新的令牌访问

					servertest.Get(a, "http://localhost:8080/info").
						Header(header.Authorization, auth.BuildToken(auth.Bearer, resp2.AccessToken)).
						Do(nil).
						Status(http.StatusOK)

					servertest.Post(a, "http://localhost:8080/refresh", nil).
						Header(header.Authorization, auth.BuildToken(auth.Bearer, resp2.RefreshToken)).
						Do(nil).
						Status(http.StatusOK).
						BodyFunc(func(a *assert.Assertion, body []byte) {
							resp3 := &Response{}
							a.NotError(json.Unmarshal(body, resp3))

							// 删除了该用户的登录信息

							servertest.Get(a, "http://localhost:8080/info").
								Header(header.Authorization, auth.BuildToken(auth.Bearer, resp3.AccessToken)).
								Do(nil).
								Status(http.StatusOK)

							token.Delete(v{ID: "5"})
							servertest.Get(a, "http://localhost:8080/info").
								Header(header.Authorization, auth.BuildToken(auth.Bearer, resp3.AccessToken)).
								Do(nil).
								Status(http.StatusUnauthorized)
						})

				})
		})
}
