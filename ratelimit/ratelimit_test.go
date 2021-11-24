// SPDX-License-Identifier: MIT

package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/issue9/assert/v2"
	"github.com/issue9/cache/memory"
)

var h1 = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
})

var _ GenFunc = GenIP

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
	srv := New(memory.New(24*time.Hour), 1, 10*time.Second, GenIP, nil)
	a.NotNil(srv)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	h := srv.MiddlewareFunc(h1)
	h.ServeHTTP(w, r)
	a.Equal(w.Code, http.StatusCreated)
	a.Equal(w.Header().Get("X-Rate-Limit-Limit"), "1")
	a.Equal(w.Header().Get("X-Rate-Limit-Remaining"), "0")

	// 没有令牌可用
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/test", nil)
	h = srv.MiddlewareFunc(h1)
	h.ServeHTTP(w, r)
	a.Equal(w.Code, http.StatusTooManyRequests)
	a.Equal(w.Header().Get("X-Rate-Limit-Limit"), "1")
	a.Equal(w.Header().Get("X-Rate-Limit-Remaining"), "0")
}
