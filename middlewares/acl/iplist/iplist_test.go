// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package iplist

import (
	"net/http"
	"os"
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/logs/v7"
	"github.com/issue9/mux/v8/header"
	"github.com/issue9/web"
	"github.com/issue9/web/mimetype/json"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/servertest"

	"github.com/issue9/webuse/v7/internal/testserver"
)

var (
	_ IPLister = &white{}
	_ IPLister = &black{}
)

func TestIPLister_Set(t *testing.T) {
	a := assert.New(t, false)

	l := NewWhite()
	a.NotNil(l)

	l.Set("192.168.1.1")
	a.Length(l.List(), 1)

	l.Set("192.168.1.1", "192.168.2/*")
	a.Length(l.List(), 2).Equal(l.List()[1], "192.168.2/*")
}

func TestWhite_Middleware(t *testing.T) {
	a := assert.New(t, false)
	s := testserver.New(a)

	l := NewWhite()
	a.NotNil(l)
	l.Set("192.168.1.1")

	router := s.Routers().New("def", nil)
	router.Use(l)
	router.Get("/test", func(ctx *web.Context) web.Responser {
		return web.Created(nil, "")
	})

	defer servertest.Run(a, s)()
	defer s.Close(0)

	servertest.Get(a, "http://localhost:8080/test").
		Header(header.XForwardedFor, "192.168.1.1").
		Do(nil).
		Status(http.StatusCreated)

	servertest.Get(a, "http://localhost:8080/test").
		Header(header.XForwardedFor, "192.168.1.2").
		Do(nil).
		Status(http.StatusForbidden)

	l.Set("192.168.1.2", "192.168.1.1", "192.168.2/*")

	servertest.Get(a, "http://localhost:8080/test").
		Header(header.XForwardedFor, "192.168.1.1").
		Do(nil).
		Status(http.StatusCreated)

	servertest.Get(a, "http://localhost:8080/test").
		Header(header.XForwardedFor, "192.168.1.2").
		Do(nil).
		Status(http.StatusCreated)

	servertest.Get(a, "http://localhost:8080/test").
		Header(header.XForwardedFor, "192.168.2.2").
		Do(nil).
		Status(http.StatusCreated)
}

func TestBlack_Middleware(t *testing.T) {
	a := assert.New(t, false)

	s, err := server.NewHTTP("test", "1.0.0", &server.Options{
		HTTPServer: &http.Server{Addr: ":8080"},
		Codec:      web.NewCodec().AddMimetype(json.Mimetype, json.Marshal, json.Unmarshal, json.ProblemMimetype),
		Logs:       logs.New(logs.NewTermHandler(os.Stderr, nil)),
	})
	a.NotError(err).NotNil(s)

	l := NewBlack()
	a.NotNil(l)
	l.Set("192.168.1.1")

	router := s.Routers().New("def", nil)
	router.Use(l)
	router.Get("/test", func(ctx *web.Context) web.Responser {
		return web.Created(nil, "")
	})

	defer servertest.Run(a, s)()
	defer s.Close(0)

	servertest.Get(a, "http://localhost:8080/test").
		Header(header.XForwardedFor, "192.168.1.1").
		Do(nil).
		Status(http.StatusForbidden)

	servertest.Get(a, "http://localhost:8080/test").
		Header(header.XForwardedFor, "192.168.1.2").
		Do(nil).
		Status(http.StatusCreated)

	l.Set("192.168.1.2", "192.168.1.1", "192.168.2/*")

	servertest.Get(a, "http://localhost:8080/test").
		Header(header.XForwardedFor, "192.168.1.1").
		Do(nil).
		Status(http.StatusForbidden)

	servertest.Get(a, "http://localhost:8080/test").
		Header(header.XForwardedFor, "192.168.1.2").
		Do(nil).
		Status(http.StatusForbidden)

	servertest.Get(a, "http://localhost:8080/test").
		Header(header.XForwardedFor, "192.168.2.2").
		Do(nil).
		Status(http.StatusForbidden)
}

func TestSplitIP(t *testing.T) {
	a := assert.New(t, false)

	ip, err := splitIP("192.168.1.1")
	a.NotError(err).Equal(ip, "192.168.1.1")

	ip, err = splitIP("192.168.1.1:8080")
	a.NotError(err).Equal(ip, "192.168.1.1")

	ip, err = splitIP("[FC00:0:130F::9C0:876A:130B]:8080")
	a.NotError(err).Equal(ip, "FC00:0:130F::9C0:876A:130B")

	ip, err = splitIP("[FC00:0:130F::9C0:876A:130B]")
	a.NotError(err).Equal(ip, "FC00:0:130F::9C0:876A:130B")
}
