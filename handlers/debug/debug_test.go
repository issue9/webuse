// SPDX-FileCopyrightText: 2014-2024 caixw
//
// SPDX-License-Identifier: MIT

package debug

import (
	"net/http"
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/web/server/servertest"

	"github.com/issue9/webuse/v7/internal/testserver"
)

func TestRegister(t *testing.T) {
	a := assert.New(t, false)
	s := testserver.New(a)
	//s.Routers().New("def", nil).Any("/debug{path}", New("path", web.ProblemNotAcceptable))
	Register(s.Routers().New("def", nil), "/debug")

	defer servertest.Run(a, s)()
	defer s.Close(0)

	servertest.Get(a, "http://localhost:8080/debug/vars").Query("seconds", "10").
		Do(nil).Status(http.StatusOK).BodyNotEmpty()

	servertest.Get(a, "http://localhost:8080/debug/pprof/cmdline").Query("seconds", "10").
		Do(nil).Status(http.StatusOK).BodyNotEmpty()

	//w = httptest.NewRecorder()
	//r = rest.Get(a, "/pprof/profile").Query("seconds", "10").Request()
	//h(s.NewContext(w, r, types.NewContext()))
	//a.Equal(w.Code, http.StatusOK).NotEmpty(w.Body.String())

	servertest.Get(a, "http://localhost:8080/debug/pprof/symbol").Query("seconds", "10").
		Do(nil).Status(http.StatusOK).BodyNotEmpty()

	servertest.Get(a, "http://localhost:8080/debug/pprof/trace").Query("seconds", "10").
		Do(nil).Status(http.StatusOK).BodyNotEmpty()

	// pprof.Index
	servertest.Get(a, "http://localhost:8080/debug/pprof/heap").Query("seconds", "10").
		Do(nil).Status(http.StatusOK).BodyNotEmpty()

	servertest.Get(a, "http://localhost:8080/debug/").Query("seconds", "10").
		Do(nil).Status(http.StatusOK).BodyFunc(func(a *assert.Assertion, body []byte) {
		a.Contains(body, debugHtml)
	})

	servertest.Get(a, "http://localhost:8080//not-exits").Query("seconds", "10").
		Do(nil).Status(http.StatusNotFound)
}
