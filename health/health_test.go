// SPDX-License-Identifier: MIT

package health

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/issue9/assert"
)

func f200(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("f200"))
}

func f401(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte("f401"))
}

func f500(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("f500"))
}

func TestHealth(t *testing.T) {
	a := assert.New(t)

	mem := NewMemory(100)
	h := New(mem)
	state := mem.Get(http.MethodGet, "/")
	a.Equal(0, state.Count)

	// 第一次访问 GET /
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	h.MiddlewareFunc(f200).ServeHTTP(w, r)
	time.Sleep(500 * time.Microsecond) // 保存是异步的，等待完成
	state = mem.Get(http.MethodGet, "/")
	a.Equal(1, state.Count)

	// 第二次访问 GET /
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/", nil)
	h.MiddlewareFunc(f500).ServeHTTP(w, r)
	time.Sleep(500 * time.Microsecond) // 保存是异步的，等待完成
	state = mem.Get(http.MethodGet, "/")
	a.Equal(2, state.Count).Equal(1, state.ServerErrors).Equal(0, state.UserErrors)

	// 第一次访问 OPTIONS /
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodOptions, "/", nil)
	h.MiddlewareFunc(f200).ServeHTTP(w, r)
	time.Sleep(500 * time.Microsecond) // 保存是异步的，等待完成
	state = mem.Get(http.MethodOptions, "/")
	a.Equal(1, state.Count)

	// 第一次访问 DELETE /users
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodDelete, "/users", nil)
	h.MiddlewareFunc(f401).ServeHTTP(w, r)
	time.Sleep(500 * time.Microsecond) // 保存是异步的，等待完成
	state = mem.Get(http.MethodDelete, "/users")
	a.Equal(1, state.Count).Equal(0, state.ServerErrors).Equal(1, state.UserErrors)

	all := h.States()
	a.Equal(3, len(all))
}
