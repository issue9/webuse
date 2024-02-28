// SPDX-License-Identifier: MIT

package access

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/web"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/servertest"
)

func TestAccess(t *testing.T) {
	a := assert.New(t, false)
	w := bytes.Buffer{}
	srv, err := server.New("test", "1.0.0", &server.Options{
		Logs:       &server.Logs{Handler: server.NewTextHandler(&w), Levels: server.AllLevels(), Created: server.MilliLayout},
		HTTPServer: &http.Server{Addr: ":8080"},
		Mimetypes:  server.JSONMimetypes(),
	})
	a.NotError(err).NotNil(srv)
	defer servertest.Run(a, srv)()
	defer srv.Close(0)

	r := srv.Routers().New("def", nil)
	m := New(srv.Logs().ERROR(), "")
	a.NotNil(m)
	r.Use(m)

	wait := make(chan bool, 1)
	r.Get("/test", func(ctx *web.Context) web.Responser {
		wait <- true
		return web.Created(nil, "")
	})

	a.Zero(w.Len())
	servertest.Get(a, "http://localhost:8080/test").Do(nil).Status(http.StatusCreated)
	a.True(w.Len() > 0)
	<-wait
}
