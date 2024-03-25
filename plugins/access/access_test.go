// SPDX-FileCopyrightText: 2015-2024 caixw
//
// SPDX-License-Identifier: MIT

package access

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/web"
	"github.com/issue9/web/server/servertest"

	"github.com/issue9/webuse/v7/internal/testserver"
)

func TestAccess(t *testing.T) {
	a := assert.New(t, false)
	srv := testserver.New(a)
	defer servertest.Run(a, srv)()
	defer srv.Close(0)

	w := bytes.Buffer{}
	r := srv.Routers().New("def", nil)
	m := New(func(s string) { w.WriteString(s) }, "")
	a.NotNil(m)
	srv.Use(m)

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
