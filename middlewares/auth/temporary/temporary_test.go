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
	"github.com/issue9/web/server/servertest"

	"github.com/issue9/webuse/v7/internal/testserver"
	"github.com/issue9/webuse/v7/middlewares/auth"
)

var _ auth.Auth[string] = &Temporary[string]{}

func TestTemporary(t *testing.T) {
	a := assert.New(t, false)
	s := testserver.New(a)

	temp := New[string](s, time.Second, true, web.ProblemForbidden, web.ProblemBadRequest)
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

			// 正常访问
			servertest.Get(a, "http://localhost:8080/info").
				Header(header.Authorization, auth.BearerToken(resp.Token)).
				Do(nil).
				Status(http.StatusOK).
				StringBody(`"5"`)

			// 再次访问，令牌失效
			servertest.Get(a, "http://localhost:8080/info").
				Header(header.Authorization, auth.BearerToken(resp.Token)).
				Do(nil).
				Status(http.StatusForbidden)
		})
}
