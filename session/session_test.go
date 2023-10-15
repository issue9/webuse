// SPDX-License-Identifier: MIT

package session

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/issue9/assert/v3"
	"github.com/issue9/web"
	"github.com/issue9/web/logs"
	"github.com/issue9/web/serializer/json"
	"github.com/issue9/web/servertest"
)

type data struct {
	ID    string `query:"id"`
	Count int    `query:"count"`
}

func TestSession(t *testing.T) {
	a := assert.New(t, false)
	srv, err := web.NewServer("test", "1.0.0", &web.Options{
		HTTPServer: &http.Server{Addr: ":8080"},
		Mimetypes: []*web.Mimetype{
			{Type: "application/json", MarshalBuilder: json.BuildMarshal, Unmarshal: json.Unmarshal},
		},
		Logs: &logs.Options{Handler: logs.NewTermHandler(os.Stdout, nil), Created: logs.NanoLayout},
	})
	a.NotError(err).NotNil(srv)

	store := NewCacheStore[data](srv.Cache(), 500*time.Microsecond)
	a.NotNil(store)

	s := New(store, 60, "sesson_id", "/", "localhost", false, false)
	a.NotNil(s)
	srv.UseMiddleware(s)

	r := srv.NewRouter("default", nil)
	r.Get("/get1", func(ctx *web.Context) web.Responser {
		want := &data{}
		if resp := ctx.QueryObject(true, want, web.ProblemInternalServerError); resp != nil {
			return resp
		}

		v, err := GetValue[data](ctx)
		a.NotError(err).Equal(v, want)

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
