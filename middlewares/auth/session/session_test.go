// SPDX-FileCopyrightText: 2015-2024 caixw
//
// SPDX-License-Identifier: MIT

package session

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/issue9/assert/v4"
	"github.com/issue9/web"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/servertest"
)

type data struct {
	ID    string `query:"id"`
	Count int    `query:"count"`
}

func TestSession(t *testing.T) {
	a := assert.New(t, false)
	srv, err := server.New("test", "1.0.0", &server.Options{
		HTTPServer: &http.Server{Addr: ":8080"},
		Mimetypes:  server.JSONMimetypes(),
		Logs:       &server.Logs{Handler: server.NewTermHandler(os.Stdout, nil), Created: server.NanoLayout},
	})
	a.NotError(err).NotNil(srv)

	store := NewCacheStore[data](srv.Cache(), 500*time.Microsecond)
	a.NotNil(store)

	s := New(store, 60, "sesson_id", "/", "localhost", false, false)
	a.NotNil(s)
	srv.Routers().Use(s)

	r := srv.Routers().New("default", nil)
	r.Get("/get1", func(ctx *web.Context) web.Responser {
		want := &data{}
		if resp := ctx.QueryObject(true, want, web.ProblemInternalServerError); resp != nil {
			return resp
		}

		sid, v, err := GetValue[data](ctx)
		a.NotError(err).
			Equal(v, want).
			NotEmpty(sid)

		v.Count++
		a.NotError(SetValue(ctx, v))

		return web.OK(nil)
	})

	defer servertest.Run(a, srv)()
	defer srv.Close(0)

	resp := servertest.Get(a, "http://localhost:8080/get1?count=0&id=").
		Do(nil).
		Status(http.StatusOK).
		Resp()
	cookie := resp.Cookies()[0]
	servertest.Get(a, "http://localhost:8080/get1?count=1&id=").
		Cookie(cookie).
		Do(nil).
		Status(http.StatusOK)
}
