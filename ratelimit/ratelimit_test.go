// SPDX-License-Identifier: MIT

package ratelimit

import (
	"net/http"
	"testing"
	"time"

	"github.com/issue9/assert/v3"
	"github.com/issue9/web"
	"github.com/issue9/web/cache"
	"github.com/issue9/web/serializer/json"
	"github.com/issue9/web/servertest"
)

var (
	_ web.Middleware = &ratelimit{}
	_ GenFunc        = GenIP
)

func TestRatelimit_Middleware(t *testing.T) {
	a := assert.New(t, false)
	s, err := web.NewServer("test", "1.0.0", &web.Options{
		HTTPServer: &http.Server{Addr: ":8080"},
		Mimetypes: []*web.Mimetype{
			{Type: "application/json", MarshalBuilder: json.BuildMarshal, Unmarshal: json.Unmarshal},
		},
	})
	a.NotError(err).NotNil(s)
	// 由 gen 方法限定在同一个请求
	srv := New(cache.Prefix(s.Cache(), "rl-"), 4, 10*time.Second, func(*web.Context) (string, error) { return "1", nil })
	a.NotNil(srv)

	r := s.NewRouter("def", nil)
	r.Use(srv)
	r.Get("/test", func(*web.Context) web.Responser {
		return web.Created(nil, "")
	})

	defer servertest.Run(a, s)()
	defer s.Close(0)

	servertest.Get(a, "http://localhost:8080/test").Do(nil).
		Status(http.StatusCreated).
		Header("X-Rate-Limit-Limit", "4").
		Header("X-Rate-Limit-Remaining", "3")

	servertest.Get(a, "http://localhost:8080/test").Do(nil).
		Status(http.StatusCreated).
		Header("X-Rate-Limit-Limit", "4").
		Header("X-Rate-Limit-Remaining", "2")

	servertest.Get(a, "http://localhost:8080/test").Do(nil).
		Status(http.StatusCreated).
		Header("X-Rate-Limit-Limit", "4").
		Header("X-Rate-Limit-Remaining", "1")

	servertest.Get(a, "http://localhost:8080/test").Do(nil).
		Status(http.StatusTooManyRequests).
		Header("X-Rate-Limit-Limit", "4").
		Header("X-Rate-Limit-Remaining", "0")

	servertest.Get(a, "http://localhost:8080/test").Do(nil).
		Status(http.StatusTooManyRequests).
		Header("X-Rate-Limit-Limit", "4").
		Header("X-Rate-Limit-Remaining", "0")
}
