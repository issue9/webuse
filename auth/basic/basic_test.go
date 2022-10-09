// SPDX-License-Identifier: MIT

package basic

import (
	"net/http"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/web"
	"github.com/issue9/web/server/servertest"
)

var (
	authFunc = func(username, password []byte) ([]byte, bool) {
		return username, true
	}

	_ web.Middleware = &Basic[[]byte]{}
)

func TestNew(t *testing.T) {
	a := assert.New(t, false)
	var b *Basic[[]byte]
	srv := servertest.NewServer(a, nil)

	a.Panic(func() {
		b = New[[]byte](srv, nil, "", false)
	})

	b = New[[]byte](srv, authFunc, "", false)

	a.Equal(b.authorization, "Authorization").
		Equal(b.authenticate, "WWW-Authenticate").
		Equal(b.unauthorizationStatus, http.StatusUnauthorized).
		NotNil(b.auth)

	b = New[[]byte](srv, authFunc, "", true)

	a.Equal(b.authorization, "Proxy-Authorization").
		Equal(b.authenticate, "Proxy-Authenticate").
		Equal(b.unauthorizationStatus, http.StatusProxyAuthRequired).
		NotNil(b.auth)
}

func TestServeHTTP_ok(t *testing.T) {
	a := assert.New(t, false)
	s := servertest.NewServer(a, nil)

	b := New[[]byte](s, authFunc, "example.com", false)
	a.NotNil(b)

	srv := servertest.NewTester(a, nil)
	r := srv.Router()
	r.Use(b)
	r.Get("/path", func(ctx *web.Context) web.Responser {
		username, found := b.GetValue(ctx)
		a.True(found).Equal(string(username), "Aladdin")
		return web.Status(http.StatusCreated)
	})

	srv.GoServe()

	srv.Get("/path").
		Do(nil).
		Header("WWW-Authenticate", `Basic realm="example.com"`).
		Status(http.StatusUnauthorized)

	// 正确的访问
	srv.Get("http://localhost:8080/path").
		Header("Authorization", "Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ=="). // Aladdin, open sesame，来自 https://zh.wikipedia.org/wiki/HTTP基本认证
		Do(nil).
		Status(http.StatusCreated)

	srv.Close(0)
}

func TestServeHTTP_failed(t *testing.T) {
	a := assert.New(t, false)
	s := servertest.NewServer(a, nil)

	b := New[[]byte](s, authFunc, "example.com", false)
	a.NotNil(b)

	srv := servertest.NewTester(a, nil)
	r := srv.Router()
	r.Use(b)
	r.Get("/path", func(ctx *web.Context) web.Responser {
		obj, found := b.GetValue(ctx)
		a.True(found).Nil(obj)
		return nil

	})

	srv.GoServe()

	srv.Get("/path").
		Do(nil).
		Header("WWW-Authenticate", `Basic realm="example.com"`).
		Status(http.StatusUnauthorized)

	// 错误的编码
	srv.Get("/path").
		Header("Authorization", "Basic aaQWxhZGRpbjpvcGVuIHNlc2FtZQ===").
		Do(nil).
		Status(http.StatusUnauthorized)

	srv.Close(0)
}
