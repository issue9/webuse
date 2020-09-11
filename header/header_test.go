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

var h1 = http.HandlerFunc(f1)

func TestNew(t *testing.T) {
	h := New(h1, map[string]string{"Server": "s1"}, nil)
	srv := rest.NewServer(t, h, nil)
	srv.NewRequest(http.MethodGet, "/test").
		Do().
		Status(http.StatusAccepted).
		Header("Server", "s1")

	// 动态生成的内容
	now := time.Now().Format("2006-01-02 15:16:05")
	h = New(h1, nil, func(h http.Header) { h.Set("Server", now) })
	srv = rest.NewServer(t, h, nil)
	srv.NewRequest(http.MethodGet, "/test").
		Do().
		Status(http.StatusAccepted).
		Header("Server", now)

		// 同时存在，则以动态生成的优先
	h = New(h1, map[string]string{"Server": "test"}, func(h http.Header) { h.Set("Server", now) })
	srv = rest.NewServer(t, h, nil)
	srv.NewRequest(http.MethodGet, "/test").
		Do().
		Status(http.StatusAccepted).
		Header("Server", now)
}
