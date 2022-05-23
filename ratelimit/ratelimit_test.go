// SPDX-License-Identifier: MIT

package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/servertest"
)

var (
	_ server.Middleware = &Ratelimit{}
	_ GenFunc           = GenIP
)

func TestGenIP(t *testing.T) {
	a := assert.New(t, false)
	s := servertest.NewServer(a, nil)
	ip4 := "1.1.1.1"
	ip6 := "[::0]"

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = ip4
	ip, err := GenIP(s.NewContext(w, r))
	a.NotError(err).Equal(ip, ip4)

	ip4 += ":8080"
	r.RemoteAddr = ip4
	ip, err = GenIP(s.NewContext(w, r))
	a.NotError(err).Equal(ip, ip4)

	r.RemoteAddr = ip6
	ip, err = GenIP(s.NewContext(w, r))
	a.NotError(err).Equal(ip, ip6)

	ip6 += ":8080"
	r.RemoteAddr = ip6
	ip, err = GenIP(s.NewContext(w, r))
	a.NotError(err).Equal(ip, ip6)

	r.RemoteAddr = ""
	ip, err = GenIP(s.NewContext(w, r))
	a.ErrorString(err, "ratelimit: 无法为请求生成唯一标记！").Zero(ip)
}

func TestRatelimit_bucket(t *testing.T) {
	a := assert.New(t, false)
	s := servertest.NewServer(a, nil)
	srv := New(s, "rl", 10, 50*time.Second, nil)
	a.NotNil(srv)

	r1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	r1.RemoteAddr = "1"
	r2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	r2.RemoteAddr = "1"

	w := httptest.NewRecorder()
	b1, err := srv.bucket(s.NewContext(w, r1))
	a.NotError(err).NotNil(b1)

	b2, err := srv.bucket(s.NewContext(w, r2))
	a.NotError(err).NotNil(b2)
	a.True(b1 == b2) // 同一个对象

	r2.RemoteAddr = "2"
	b2, err = srv.bucket(s.NewContext(w, r2))
	a.NotError(err).NotNil(b2)
	a.False(b1 == b2) // 不同对象

	r2.RemoteAddr = ""
	b2, err = srv.bucket(s.NewContext(w, r2))
	a.Error(err).Nil(b2)
}

func TestRatelimit_Transfer(t *testing.T) {
	a := assert.New(t, false)
	s := servertest.NewServer(a, nil)
	srv := New(s, "rl", 10, 50*time.Second, nil)
	a.NotNil(srv)

	r1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	r1.RemoteAddr = "1"
	r2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	r2.RemoteAddr = "2"

	w := httptest.NewRecorder()
	b1, err := srv.bucket(s.NewContext(w, r1))
	a.NotError(err).NotNil(b1)
	b2, err := srv.bucket(s.NewContext(w, r2))
	a.NotError(err).NotNil(b2)
	a.NotEmpty(b1, b2)

	a.NotError(srv.Transfer("1", "2"))
	b2, err = srv.bucket(s.NewContext(w, r2))
	a.NotError(err).NotNil(b2)
	a.Equal(b1, b2)
}

func TestRatelimit_Middleware(t *testing.T) {
	a := assert.New(t, false)
	s := servertest.NewServer(a, nil)
	srv := New(s, "rl", 1, 10*time.Second, GenIP)
	a.NotNil(srv)

	w := httptest.NewRecorder()
	r := rest.Get(a, "/test").Header("accept", "text/plain").Request()
	r.RemoteAddr = "192.168.1.1"
	ctx := s.NewContext(w, r)
	srv.Middleware(servertest.BuildHandler(http.StatusCreated))(ctx).Apply(ctx)
	a.Equal(w.Code, http.StatusCreated).
		Equal(w.Body.String(), []byte("201"))
	a.Equal(w.Header().Get("X-Rate-Limit-Limit"), "1")
	a.Equal(w.Header().Get("X-Rate-Limit-Remaining"), "0")

	// 没有令牌可用
	w = httptest.NewRecorder()
	r = rest.Get(a, "/test").Header("accept", "text/plain").Request()
	r.RemoteAddr = "192.168.1.1"
	ctx = s.NewContext(w, r)
	srv.Middleware(servertest.BuildHandler(http.StatusTooManyRequests))(ctx).Apply(ctx)
	a.Equal(w.Code, http.StatusTooManyRequests).
		Zero(w.Body.Len())
	a.Equal(w.Header().Get("X-Rate-Limit-Limit"), "1")
	a.Equal(w.Header().Get("X-Rate-Limit-Remaining"), "0")
}
