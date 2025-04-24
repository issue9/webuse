// SPDX-FileCopyrightText: 2025 caixw
//
// SPDX-License-Identifier: MIT

package mimetype

import (
	"net/http"
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/mux/v9/header"
	"github.com/issue9/web"
	"github.com/issue9/web/server/servertest"

	"github.com/issue9/webuse/v7/internal/testserver"
)

func TestMimetype(t *testing.T) {
	a := assert.New(t, false)
	s := testserver.New(a)

	next := func(ctx *web.Context) web.Responser {
		return web.Created(nil, "")
	}

	router := s.Routers().New("def", nil)
	router.Any("/test", next, New("application/json", "application/cbor"))

	defer servertest.Run(a, s)()
	defer s.Close(0)

	servertest.Get(a, "http://localhost:8080/test").
		Header(header.Accept, header.JSON).
		Do(nil).
		Status(http.StatusCreated)

	servertest.Post(a, "http://localhost:8080/test", nil).
		Header(header.Accept, header.JSON+";charset=utf-8").
		Do(nil).
		Status(http.StatusCreated)

	servertest.Patch(a, "http://localhost:8080/test", nil).
		Header(header.Accept, header.Javascript+";charset=utf-8").
		Do(nil).
		Status(http.StatusNotAcceptable)

	servertest.NewRequest(a, http.MethodOptions, "http://localhost:8080/test").
		Header(header.Accept, header.JSON+";charset=utf-8").
		Do(nil).
		Status(http.StatusOK).
		Header(header.Accept, header.JSON)
}
