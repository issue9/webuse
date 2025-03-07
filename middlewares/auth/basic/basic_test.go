// SPDX-FileCopyrightText: 2015-2025 caixw
//
// SPDX-License-Identifier: MIT

package basic

import (
	"net/http"
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/mux/v9/header"
	"github.com/issue9/web"
	"github.com/issue9/web/server/servertest"

	"github.com/issue9/webuse/v7/internal/testserver"
	"github.com/issue9/webuse/v7/middlewares/auth"
)

var (
	authFunc = func(username, password []byte) ([]byte, bool) {
		return username, true
	}

	_ auth.Auth[[]byte] = &basic[[]byte]{}
)

func TestNew(t *testing.T) {
	a := assert.New(t, false)
	srv := testserver.New(a)
	var b *basic[[]byte]

	a.Panic(func() {
		New[[]byte](srv, nil, "", false)
	})

	b = New(srv, authFunc, "", false).(*basic[[]byte])

	a.Equal(b.authorization, header.Authorization).
		Equal(b.authenticate, header.WWWAuthenticate).
		Equal(b.problemID, web.ProblemUnauthorized).
		NotNil(b.auth)

	b = New(srv, authFunc, "", true).(*basic[[]byte])

	a.Equal(b.authorization, header.ProxyAuthorization).
		Equal(b.authenticate, header.ProxyAuthenticate).
		Equal(b.problemID, web.ProblemProxyAuthRequired).
		NotNil(b.auth)
}

func TestServeHTTP_ok(t *testing.T) {
	a := assert.New(t, false)
	s := testserver.New(a)

	b := New(s, authFunc, "example.com", false)
	a.NotNil(b)

	r := s.Routers().New("def", nil)
	r.Use(b)
	r.Get("/path", func(ctx *web.Context) web.Responser {
		username, found := b.GetInfo(ctx)
		a.True(found).Equal(string(username), "Aladdin")
		return web.Status(http.StatusCreated)
	})

	defer servertest.Run(a, s)()
	defer s.Close(0)

	servertest.Get(a, "http://localhost:8080/path").
		Do(nil).
		Header(header.WWWAuthenticate, auth.BasicToken(`realm="example.com"`)).
		Status(http.StatusUnauthorized)

	// 正确的访问
	servertest.Get(a, "http://localhost:8080/path").
		Header(header.Authorization, auth.BasicToken("QWxhZGRpbjpvcGVuIHNlc2FtZQ==")). // Aladdin, open sesame，来自 https://zh.wikipedia.org/wiki/HTTP基本认证
		Do(nil).
		Status(http.StatusCreated)
}

func TestServeHTTP_failed(t *testing.T) {
	a := assert.New(t, false)
	s := testserver.New(a)

	b := New(s, authFunc, "example.com", false)
	a.NotNil(b)

	r := s.Routers().New("def", nil)
	r.Use(b)
	r.Get("/path", func(ctx *web.Context) web.Responser {
		obj, found := b.GetInfo(ctx)
		a.True(found).Nil(obj)
		return nil
	})

	defer servertest.Run(a, s)()
	defer s.Close(0)

	servertest.Get(a, "http://localhost:8080/path").
		Do(nil).
		Header(header.WWWAuthenticate, auth.BasicToken(`realm="example.com"`)).
		Status(http.StatusUnauthorized)

	// 错误的编码
	servertest.Get(a, "http://localhost:8080/path").
		Header(header.Authorization, auth.BasicToken("aaQWxhZGRpbjpvcGVuIHNlc2FtZQ===")).
		Do(nil).
		Status(http.StatusUnauthorized)
}
