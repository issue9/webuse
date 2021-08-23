// SPDX-License-Identifier: MIT

package recovery

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

var f1 = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
}

var h1 = http.HandlerFunc(f1)

func TestDefaultRecoverFunc(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()
	a.NotNil(w)

	DefaultRecoverFunc(http.StatusInternalServerError)(w, "not found")
	a.Equal(http.StatusText(http.StatusInternalServerError)+"\n", w.Body.String())
}

func TestRecoverFunc_Middleware(t *testing.T) {
	a := assert.New(t)

	// DefaultRecoverFunc
	h := DefaultRecoverFunc(http.StatusInternalServerError)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "http://caixw.io/test", nil)
	a.NotNil(h).NotNil(w).NotNil(r)
	h.Middleware(h1).ServeHTTP(w, r)
	a.Equal(w.Code, http.StatusCreated)

	// 触发 panic
	next := func(w http.ResponseWriter, r *http.Request) {
		panic("test")
	}
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "http://caixw.io/test", nil)
	a.NotNil(h).NotNil(w).NotNil(r)
	h.MiddlewareFunc(next).ServeHTTP(w, r)
	a.Equal(w.Code, http.StatusInternalServerError)
}

func TestTraceStack(t *testing.T) {
	w := httptest.NewRecorder()
	TraceStack(http.StatusNotFound)(w, "TraceStack")
	t.Log(w.Body.String())
}
