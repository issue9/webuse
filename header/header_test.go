// SPDX-License-Identifier: MIT

package header

import (
	"net/http"
	"testing"
	"time"

	"github.com/issue9/assert/rest"
)

var f1 = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
}

func TestNew(t *testing.T) {
	h := New(map[string]string{"Server": "s1"}, nil)
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

	// 动态生成的内容
	now := time.Now().Format("2006-01-02 15:16:05")
	h = New(nil, func(h http.Header) { h.Set("Server", now) })
	srv = rest.NewServer(t, h.MiddlewareFunc(f1), nil)
	srv.NewRequest(http.MethodGet, "/test").
		Do().
		Status(http.StatusAccepted).
		Header("Server", now)

	// 同时存在，则以动态生成的优先
	h = New(map[string]string{"Server": "test"}, func(h http.Header) { h.Set("Server", now) })
	srv = rest.NewServer(t, h.MiddlewareFunc(f1), nil)
	srv.NewRequest(http.MethodGet, "/test").
		Do().
		Status(http.StatusAccepted).
		Header("Server", now)
}
