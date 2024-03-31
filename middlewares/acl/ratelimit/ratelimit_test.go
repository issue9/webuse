// SPDX-FileCopyrightText: 2015-2024 caixw
//
// SPDX-License-Identifier: MIT

package ratelimit

import (
	"net/http"
	"testing"
	"time"

	"github.com/issue9/assert/v4"
	"github.com/issue9/cache"
	"github.com/issue9/mux/v8/header"
	"github.com/issue9/web"
	"github.com/issue9/web/server/servertest"

	"github.com/issue9/webuse/v7/internal/testserver"
)

var _ web.Middleware = &ratelimit{}

func TestRatelimit_Middleware(t *testing.T) {
	a := assert.New(t, false)
	s := testserver.New(a)

	// 由 gen 方法限定在同一个请求
	srv := New(cache.Prefix(s.Cache(), "rl-"), 4, 10*time.Second, func(*web.Context) (string, error) { return "1", nil }, nil)
	a.NotNil(srv)

	r := s.Routers().New("def", nil)
	r.Use(srv)
	r.Get("/test", func(*web.Context) web.Responser {
		return web.Created(nil, "")
	})

	defer servertest.Run(a, s)()
	defer s.Close(0)

	servertest.Get(a, "http://localhost:8080/test").Do(nil).
		Status(http.StatusCreated).
		Header(header.XRateLimitLimit, "4").
		Header(header.XRateLimitRemaining, "3")

	servertest.Get(a, "http://localhost:8080/test").Do(nil).
		Status(http.StatusCreated).
		Header(header.XRateLimitLimit, "4").
		Header(header.XRateLimitRemaining, "2")

	servertest.Get(a, "http://localhost:8080/test").Do(nil).
		Status(http.StatusCreated).
		Header(header.XRateLimitLimit, "4").
		Header(header.XRateLimitRemaining, "1")

	servertest.Get(a, "http://localhost:8080/test").Do(nil).
		Status(http.StatusTooManyRequests).
		Header(header.XRateLimitLimit, "4").
		Header(header.XRateLimitRemaining, "0")

	servertest.Get(a, "http://localhost:8080/test").Do(nil).
		Status(http.StatusTooManyRequests).
		Header(header.XRateLimitLimit, "4").
		Header(header.XRateLimitRemaining, "0")
}
