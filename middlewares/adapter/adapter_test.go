// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package adapter

import (
	"net/http"
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/web/server/servertest"

	"github.com/issue9/webuse/v7/internal/testserver"
)

func TestAdapter(t *testing.T) {
	a := assert.New(t, false)
	s := testserver.New(a)

	r := s.Routers().New("def", nil)
	r.Get("/get", HTTPHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))

	defer servertest.Run(a, s)()
	defer s.Close(0)

	servertest.Get(a, "http://localhost:8080/get").Do(nil).Status(http.StatusAccepted)
}
