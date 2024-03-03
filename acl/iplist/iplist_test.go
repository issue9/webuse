// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package iplist

import (
	"net/http"
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/web"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/servertest"
)

var _ web.Middleware = &IPList{}

func TestIPList_WithWhite(t *testing.T) {
	a := assert.New(t, false)

	l := New([]string{"192.168.1.1"}, nil)
	a.NotNil(l).
		Length(l.white, 1).
		Length(l.whiteWildcard, 0)

	l.WithWhite("192.168.1.1")
	a.Length(l.white, 1).
		Length(l.whiteWildcard, 0)

	l.WithWhite("192.168.1/*")
	a.Length(l.white, 1).
		Length(l.whiteWildcard, 1)
}

func TestIPList_WithBlack(t *testing.T) {
	a := assert.New(t, false)

	l := New([]string{"192.168.1.1"}, nil)
	a.NotNil(l).
		Length(l.black, 0).
		Length(l.blackWildcard, 0)

	l.WithBlack("192.168.1.1")
	a.Length(l.black, 1).
		Length(l.blackWildcard, 0)

	l.WithBlack("192.168.1/*")
	a.Length(l.black, 1).
		Length(l.blackWildcard, 1)
}

func TestIPList_Middleware(t *testing.T) {
	a := assert.New(t, false)

	s, err := server.New("test", "1.0.0", &server.Options{
		HTTPServer: &http.Server{Addr: ":8080"},
		Mimetypes:  server.JSONMimetypes(),
	})
	a.NotError(err).NotNil(s)

	l := New([]string{"192.168.1.1"}, nil)
	a.NotNil(l)

	router := s.Routers().New("def", nil)
	router.Use(l)
	router.Get("/test", func(ctx *web.Context) web.Responser {
		return web.Created(nil, "")
	})

	defer servertest.Run(a, s)()
	defer s.Close(0)

	servertest.Get(a, "http://localhost:8080/test").
		Header("X-Forwarded-For", "192.168.1.1").
		Do(nil).
		Status(http.StatusCreated)

	servertest.Get(a, "http://localhost:8080/test").
		Header("X-Forwarded-For", "192.168.1.2").
		Do(nil).
		Status(http.StatusCreated)

	l.WithBlack("192.168.1.2")
	l.WithBlack("192.168.1.1")
	l.WithBlack("192.168.2/*")

	// 同时存在于黑名单和白名单
	servertest.Get(a, "http://localhost:8080/test").
		Header("X-Forwarded-For", "192.168.1.1").
		Do(nil).
		Status(http.StatusCreated)

	// 在黑名单中
	servertest.Get(a, "http://localhost:8080/test").
		Header("X-Forwarded-For", "192.168.1.2").
		Do(nil).
		Status(http.StatusForbidden)

	servertest.Get(a, "http://localhost:8080/test").
		Header("X-Forwarded-For", "192.168.2.2").
		Do(nil).
		Status(http.StatusForbidden)
}

func TestIsPort(t *testing.T) {
	a := assert.New(t, false)

	a.True(isPort("333")).
		False(isPort(":333")).
		False(isPort(":]333"))
}
