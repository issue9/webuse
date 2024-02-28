// SPDX-License-Identifier: MIT

package basic

import (
	"net/http"
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/web"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/servertest"
)

var (
	authFunc = func(username, password []byte) ([]byte, bool) {
		return username, true
	}

	_ web.Middleware = &basic[[]byte]{}
)

func TestNew(t *testing.T) {
	a := assert.New(t, false)
	var b *basic[[]byte]
	srv, err := server.New("test", "1.0.0", &server.Options{
		HTTPServer: &http.Server{Addr: ":8080"},
		Mimetypes:  server.JSONMimetypes(),
	})
	a.NotError(err).NotNil(srv)

	a.Panic(func() {
		New[[]byte](srv, nil, "", false)
	})

	b = New(srv, authFunc, "", false).(*basic[[]byte])

	a.Equal(b.authorization, "Authorization").
		Equal(b.authenticate, "WWW-Authenticate").
		Equal(b.problemID, web.ProblemUnauthorized).
		NotNil(b.auth)

	b = New(srv, authFunc, "", true).(*basic[[]byte])

	a.Equal(b.authorization, "Proxy-Authorization").
		Equal(b.authenticate, "Proxy-Authenticate").
		Equal(b.problemID, web.ProblemProxyAuthRequired).
		NotNil(b.auth)
}

func TestServeHTTP_ok(t *testing.T) {
	a := assert.New(t, false)
	s, err := server.New("test", "1.0.0", &server.Options{
		HTTPServer: &http.Server{Addr: ":8080"},
		Mimetypes:  server.JSONMimetypes(),
	})
	a.NotError(err).NotNil(s)

	b := New(s, authFunc, "example.com", false)
	a.NotNil(b)

	r := s.Routers().New("def", nil)
	r.Use(b)
	r.Get("/path", func(ctx *web.Context) web.Responser {
		username, found := GetValue[[]byte](ctx)
		a.True(found).Equal(string(username), "Aladdin")
		return web.Status(http.StatusCreated)
	})

	defer servertest.Run(a, s)()
	defer s.Close(0)

	servertest.Get(a, "http://localhost:8080/path").
		Do(nil).
		Header("WWW-Authenticate", `Basic realm="example.com"`).
		Status(http.StatusUnauthorized)

	// 正确的访问
	servertest.Get(a, "http://localhost:8080/path").
		Header("Authorization", "Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ=="). // Aladdin, open sesame，来自 https://zh.wikipedia.org/wiki/HTTP基本认证
		Do(nil).
		Status(http.StatusCreated)
}

func TestServeHTTP_failed(t *testing.T) {
	a := assert.New(t, false)
	s, err := server.New("test", "1.0.0", &server.Options{
		HTTPServer: &http.Server{Addr: ":8080"},
		Mimetypes:  server.JSONMimetypes(),
	})
	a.NotError(err).NotNil(s)

	b := New(s, authFunc, "example.com", false)
	a.NotNil(b)

	r := s.Routers().New("def", nil)
	r.Use(b)
	r.Get("/path", func(ctx *web.Context) web.Responser {
		obj, found := GetValue[[]byte](ctx)
		a.True(found).Nil(obj)
		return nil
	})

	defer servertest.Run(a, s)()
	defer s.Close(0)

	servertest.Get(a, "http://localhost:8080/path").
		Do(nil).
		Header("WWW-Authenticate", `Basic realm="example.com"`).
		Status(http.StatusUnauthorized)

	// 错误的编码
	servertest.Get(a, "http://localhost:8080/path").
		Header("Authorization", "Basic aaQWxhZGRpbjpvcGVuIHNlc2FtZQ===").
		Do(nil).
		Status(http.StatusUnauthorized)
}
