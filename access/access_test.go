// SPDX-License-Identifier: MIT

package access

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/web"
	"github.com/issue9/web/logs"
	"github.com/issue9/web/server/servertest"
)

func TestAccess(t *testing.T) {
	a := assert.New(t, false)
	w := bytes.Buffer{}
	srv := servertest.NewTester(a, &web.Options{
		Logs: &logs.Options{Writer: logs.NewTextWriter(logs.MilliLayout, &w)},
	})

	r := srv.Server().Routers().New("def", nil)
	m := New(srv.Server().Logs().ERROR(), "")
	a.NotNil(m)
	r.Use(m)

	wait := make(chan bool, 1)
	r.Get("/test", func(ctx *web.Context) web.Responser {
		wait <- true
		return web.Created(nil, "")
	})

	srv.GoServe()

	a.Zero(w.Len())

	srv.Get("/test").Do(nil).Status(http.StatusCreated)
	<-wait
	a.True(w.Len() > 0)

	srv.Close(0)
}
