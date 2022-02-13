// SPDX-License-Identifier: MIT

package basic

import (
	"io/ioutil"
	"log"
	"net/http"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/web"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/servertest"

	"github.com/issue9/middleware/v6/auth"
)

var (
	authFunc = func(username, password []byte) (interface{}, bool) {
		return username, true
	}

	_ server.Middleware = &Basic{}
)

func TestNew(t *testing.T) {
	a := assert.New(t, false)
	var b *Basic

	a.Panic(func() {
		b = New(nil, "", false, nil)
	})

	a.NotPanic(func() {
		b = New(authFunc, "", false, nil)
	})

	a.Equal(b.authorization, "Authorization").
		Equal(b.authenticate, "WWW-Authenticate").
		Equal(b.unauthorizationStatus, http.StatusUnauthorized).
		Nil(b.errlog).
		NotNil(b.auth)

	a.NotPanic(func() {
		b = New(authFunc, "", true, log.New(ioutil.Discard, "", 0))
	})

	a.Equal(b.authorization, "Proxy-Authorization").
		Equal(b.authenticate, "Proxy-Authenticate").
		Equal(b.unauthorizationStatus, http.StatusProxyAuthRequired).
		NotNil(b.errlog).
		NotNil(b.auth)
}

func TestServeHTTP_ok(t *testing.T) {
	a := assert.New(t, false)
	b := New(authFunc, "example.com", false, nil)
	a.NotNil(b)

	srv := servertest.NewTester(a, nil)
	r := srv.NewRouter(b)
	r.Get("/path", func(ctx *web.Context) *web.Response {
		username, found := auth.GetValue(ctx)
		a.True(found).Equal(string(username.([]byte)), "Aladdin")
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
	srv.Wait()
}

func TestServeHTTP_failed(t *testing.T) {
	a := assert.New(t, false)
	b := New(authFunc, "example.com", false, nil)
	a.NotNil(b)

	srv := servertest.NewTester(a, nil)
	r := srv.NewRouter(b)
	r.Get("/path", func(ctx *web.Context) *web.Response {
		obj, found := auth.GetValue(ctx)
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
}
