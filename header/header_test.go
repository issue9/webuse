// SPDX-License-Identifier: MIT

package header

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"
)

var f1 = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
}

func TestNew(t *testing.T) {
	a := assert.New(t)

	h := New(nil)
	a.NotPanic(func() {
		h.Set("key", "val")
	})

	h = New(map[string]string{"Server": "s1"})
	srv := rest.NewServer(t, h.MiddlewareFunc(f1), nil)
	srv.NewRequest(http.MethodGet, "/test").
		Do().
		Status(http.StatusAccepted).
		Header("Server", "s1").
		Header("Content-Type", "")

	// Set
	h.Set("Server", "s2").
		Set("Content-Type", "xml")
	srv = rest.NewServer(t, h.MiddlewareFunc(f1), nil)
	srv.NewRequest(http.MethodGet, "/test").
		Do().
		Status(http.StatusAccepted).
		Header("Server", "s2").
		Header("Content-Type", "xml")

	// Delete
	h.Delete("Server")
	srv = rest.NewServer(t, h.MiddlewareFunc(f1), nil)
	srv.NewRequest(http.MethodGet, "/test").
		Do().
		Status(http.StatusAccepted).
		Header("Server", "").
		Header("Content-Type", "xml")
}
