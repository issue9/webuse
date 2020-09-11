// SPDX-License-Identifier: MIT

package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/issue9/assert"
)

var _ GenFunc = GenIP

func TestGenIP(t *testing.T) {
	a := assert.New(t)
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

func TestServer_bucket(t *testing.T) {
	a := assert.New(t)
	srv := NewServer(NewMemory(10), 10, 50*time.Second, nil)
	a.NotNil(srv)

	r1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	r1.RemoteAddr = "1"
	r2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	r2.RemoteAddr = "1"

	b1, err := srv.bucket(r1)
	a.NotError(err).NotNil(b1)
	b2, err := srv.bucket(r2)
	a.NotError(err).NotNil(b2)
	a.Equal(b1, b2)

	r2.RemoteAddr = "2"
	b2, err = srv.bucket(r2)
	a.NotError(err).NotNil(b2)
	a.NotEqual(b1, b2)

	r2.RemoteAddr = ""
	b2, err = srv.bucket(r2)
	a.Error(err).Nil(b2)
}

func TestServer_transfer(t *testing.T) {
	a := assert.New(t)
	srv := NewServer(NewMemory(10), 10, 50*time.Second, nil)
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
