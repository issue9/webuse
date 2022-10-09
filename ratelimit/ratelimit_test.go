// SPDX-License-Identifier: MIT

package ratelimit

import (
	"net/http"
	"testing"
	"time"

	"github.com/issue9/assert/v3"
	"github.com/issue9/web"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/servertest"
)

var (
	_ web.Middleware = &Ratelimit{}
	_ GenFunc        = GenIP
)

func TestRatelimit_Middleware(t *testing.T) {
	a := assert.New(t, false)
	s := servertest.NewTester(a, nil)
	// 由 gen 方法限定在同一个请求
	srv := New(s.Server(), "rl", 3, 10*time.Second, func(*web.Context) (string, error) { return "1", nil })
	a.NotNil(srv)

	r := s.Router()
	r.Use(srv)
	r.Get("/test", func(*server.Context) server.Responser {
		return web.Created(nil, "")
	})

	s.GoServe()

	s.Get("/test").Do(nil).
		Status(http.StatusCreated).
		Header("X-Rate-Limit-Limit", "3").
		Header("X-Rate-Limit-Remaining", "2")

	s.Get("/test").Do(nil).
		Status(http.StatusCreated).
		Header("X-Rate-Limit-Limit", "3").
		Header("X-Rate-Limit-Remaining", "1")

	s.Get("/test").Do(nil).
		Status(http.StatusCreated).
		Header("X-Rate-Limit-Limit", "3").
		Header("X-Rate-Limit-Remaining", "0")

	s.Get("/test").Do(nil).
		Status(http.StatusTooManyRequests).
		Header("X-Rate-Limit-Limit", "3").
		Header("X-Rate-Limit-Remaining", "0")

	s.Close(0)
}
