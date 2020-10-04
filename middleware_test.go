// SPDX-License-Identifier: MIT

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

func f1(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("f1-"))
}

var h1 = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("h1-"))
})

func m1(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("1"))
		h.ServeHTTP(w, r)
	})
}

func m2(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("2"))
		h.ServeHTTP(w, r)
	})
}

func m3(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("3"))
		h.ServeHTTP(w, r)
	})
}

func TestHandlerFunc(t *testing.T) {
	a := assert.New(t)

	h := HandlerFunc(f1, m1, m2, m3)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	h.ServeHTTP(w, r)
	a.Equal(w.Body.String(), "123f1-")

	// 未指定 middleware

	h = HandlerFunc(f1)
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	h.ServeHTTP(w, r)
	a.Equal(w.Body.String(), "f1-")
}
