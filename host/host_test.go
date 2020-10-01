// SPDX-License-Identifier: MIT

package host

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

var f1 = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
}

var h1 = http.HandlerFunc(f1)

func TestNew(t *testing.T) {
	a := assert.New(t)

	h := New(false, "caixw.io", "caixw.oi", "*.example.com")
	a.NotNil(h)
	a.Equal(len(h.domains), 2).
		Equal(len(h.wildcards), 1)

	// HTTP
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "http://caixw.io/test", nil)
	a.NotNil(w).NotNil(r)
	h.MiddlewareFunc(f1).ServeHTTP(w, r)
	a.Equal(w.Code, http.StatusAccepted)

	// HTTPS
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "https://caixw.io/test", nil)
	a.NotNil(w).NotNil(r)
	h.MiddlewareFunc(f1).ServeHTTP(w, r)
	a.Equal(w.Code, http.StatusAccepted)

	// 泛域名
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "https://xx.example.com/test", nil)
	a.NotNil(w).NotNil(r)
	h.MiddlewareFunc(f1).ServeHTTP(w, r)
	a.Equal(w.Code, http.StatusAccepted)

	// 带端口
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "http://caixw.io:88/test", nil)
	a.NotNil(w).NotNil(r)
	h.MiddlewareFunc(f1).ServeHTTP(w, r)
	a.Equal(w.Code, http.StatusAccepted)

	// 访问不允许的域名
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "http://sub.caixw.io/test", nil)
	h.MiddlewareFunc(f1).ServeHTTP(w, r)
	a.NotNil(w).NotNil(r)
	a.Equal(w.Code, http.StatusNotFound)

	// 访问不允许的域名
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "http://sub.1example.com/test", nil)
	h.MiddlewareFunc(f1).ServeHTTP(w, r)
	a.NotNil(w).NotNil(r)
	a.Equal(w.Code, http.StatusNotFound)
}

func TestHost_Add_Delete(t *testing.T) {
	a := assert.New(t)

	h := New(false)

	h.Add("xx.example.com")
	h.Add("xx.example.com")
	h.Add("xx.example.com")
	h.Add("*.example.com")
	h.Add("*.example.com")
	h.Add("*.example.com")
	a.Equal(1, len(h.domains)).
		Equal(1, len(h.wildcards))

	h.Delete("*.example.com")
	a.Equal(1, len(h.domains)).
		Equal(0, len(h.wildcards))

	h.Delete("*.example.com")
	a.Equal(1, len(h.domains)).
		Equal(0, len(h.wildcards))

	h.Delete("xx.example.com")
	a.Equal(0, len(h.domains)).
		Equal(0, len(h.wildcards))
}

func TestNew_empty_domains(t *testing.T) {
	a := assert.New(t)

	h := New(false)
	a.NotNil(h)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "http://caixw.io/test", nil)
	a.NotNil(w).NotNil(r)
	h.MiddlewareFunc(f1).ServeHTTP(w, r)
	a.Equal(w.Code, http.StatusNotFound)

	h = New(true)
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "http://caixw.io/test", nil)
	h.MiddlewareFunc(f1).ServeHTTP(w, r)
	a.Equal(w.Code, http.StatusAccepted)

	// 访问不允许的域名
	h = New(false)
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "http://not.exsits/test", nil)
	h.MiddlewareFunc(f1).ServeHTTP(w, r)
	a.NotNil(w).NotNil(r)
	a.Equal(w.Code, http.StatusNotFound)

	h = New(true)
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "http://not.exsits/test", nil)
	h.MiddlewareFunc(f1).ServeHTTP(w, r)
	a.NotNil(w).NotNil(r)
	a.Equal(w.Code, http.StatusAccepted)
}
