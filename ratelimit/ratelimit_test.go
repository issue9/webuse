// SPDX-License-Identifier: MIT

package ratelimit

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"
	"github.com/issue9/cache/memory"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/servertest"
)

var (
	_ server.Middleware = &Ratelimit{}
	_ GenFunc           = GenIP
)

func TestGenIP(t *testing.T) {
	a := assert.New(t, false)
	ip4 := "1.1.1.1"
	ip6 := "[::0]"

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = ip4
	ip, err := GenIP(r)
	a.NotError(err).Equal(ip, ip4)

	r.RemoteAddr = ip4 + ":8080"
	ip, err = GenIP(r)
	a.NotError(err).Equal(ip, ip4)

	r.RemoteAddr = ip6
	ip, err = GenIP(r)
	a.NotError(err).Equal(ip, ip6)

	r.RemoteAddr = ip6 + ":8080"
	ip, err = GenIP(r)
	a.NotError(err).Equal(ip, ip6)
}

func TestRatelimit_bucket(t *testing.T) {
	a := assert.New(t, false)
	srv := New(memory.New(24*time.Hour), 10, 50*time.Second, nil, nil)
	a.NotNil(srv)

	r1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	r1.RemoteAddr = "1"
	r2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	r2.RemoteAddr = "1"

	b1, err := srv.bucket(r1)
	a.NotError(err).NotNil(b1)
	b2, err := srv.bucket(r2)
	a.NotError(err).NotNil(b2)
	a.True(b1 == b2) // 同一个对象

	r2.RemoteAddr = "2"
	b2, err = srv.bucket(r2)
	a.NotError(err).NotNil(b2)
	a.False(b1 == b2) // 不同对象

	r2.RemoteAddr = ""
	b2, err = srv.bucket(r2)
	a.Error(err).Nil(b2)
}

func TestRatelimit_Transfer(t *testing.T) {
	a := assert.New(t, false)
	srv := New(memory.New(24*time.Hour), 10, 50*time.Second, nil, nil)
	a.NotNil(srv)

	r1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	r1.RemoteAddr = "1"
	r2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	r2.RemoteAddr = "2"

	b1, err := srv.bucket(r1)
	a.NotError(err).NotNil(b1)
	b2, err := srv.bucket(r2)
	a.NotError(err).NotNil(b2)
	a.NotEmpty(b1, b2)

	a.NotError(srv.Transfer("1", "2"))
	b2, err = srv.bucket(r2)
	a.NotError(err).NotNil(b2)
	a.Equal(b1, b2)
}

func TestRatelimit_Middleware(t *testing.T) {
	a := assert.New(t, false)
	srv := New(memory.New(24*time.Hour), 1, 10*time.Second, GenIP, log.Default())
	a.NotNil(srv)
	s := servertest.NewServer(a, nil)

	w := httptest.NewRecorder()
	r := rest.Get(a, "/test").Request()
	r.RemoteAddr = "192.168.1.1"
	resp := srv.Middleware(servertest.BuildHandler(http.StatusCreated))(s.NewContext(w, r))
	a.Equal(resp.Status(), http.StatusCreated).
		Equal(resp.Body(), []byte("201"))
	h, found := resp.GetHeader("X-Rate-Limit-Limit")
	a.True(found).Equal(h, "1")
	h, found = resp.GetHeader("X-Rate-Limit-Remaining")
	a.True(found).Equal(h, "0")

	// 没有令牌可用
	w = httptest.NewRecorder()
	r = rest.Get(a, "/test").Request()
	r.RemoteAddr = "192.168.1.1"
	resp = srv.Middleware(servertest.BuildHandler(http.StatusTooManyRequests))(s.NewContext(w, r))
	a.Equal(resp.Status(), http.StatusTooManyRequests).
		Empty(resp.Body())
	h, found = resp.GetHeader("X-Rate-Limit-Limit")
	a.True(found).Equal(h, "1")
	h, found = resp.GetHeader("X-Rate-Limit-Remaining")
	a.True(found).Equal(h, "0")
}
