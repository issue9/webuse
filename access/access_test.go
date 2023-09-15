// SPDX-License-Identifier: MIT

package access

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/web"
	"github.com/issue9/web/logs"
	"github.com/issue9/web/serializer/json"
	"github.com/issue9/web/servertest"
)

func TestAccess(t *testing.T) {
	a := assert.New(t, false)
	w := bytes.Buffer{}
	srv, err := web.NewServer("test", "1.0.0", &web.Options{
		Logs:       &logs.Options{Handler: logs.NewTextHandler(logs.MilliLayout, &w), Levels: logs.AllLevels()},
		HTTPServer: &http.Server{Addr: ":8080"},
		Mimetypes: []*web.Mimetype{
			{Type: "application/json", Marshal: json.Marshal, Unmarshal: json.Unmarshal},
		},
	})
	a.NotError(err).NotNil(srv)
	defer servertest.Run(a, srv)()
	defer srv.Close(0)

	r := srv.NewRouter("def", nil)
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
