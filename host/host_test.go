// SPDX-License-Identifier: MIT

package host

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

var f1 = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(1)
}

var h1 = http.HandlerFunc(f1)

func TestNew(t *testing.T) {
	a := assert.New(t)

	h := New(h1, "caixw.io", "caixw.oi", "*.example.com")
	a.NotNil(h)
	hh, ok := h.(*host)
	a.True(ok).NotNil(hh)
	a.Equal(len(hh.domains), 2).
		Equal(len(hh.wildcards), 1)

	// HTTP
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "http://caixw.io/test", nil)
	a.NotNil(w).NotNil(r)
	h.ServeHTTP(w, r)
	a.Equal(w.Code, 1)

	// HTTPS
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "https://caixw.io/test", nil)
	a.NotNil(w).NotNil(r)
	h.ServeHTTP(w, r)
	a.Equal(w.Code, 1)

	// 泛域名
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "https://xx.example.com/test", nil)
	a.NotNil(w).NotNil(r)
	h.ServeHTTP(w, r)
	a.Equal(w.Code, 1)

	// 带端口
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "http://caixw.io:88/test", nil)
	a.NotNil(w).NotNil(r)
	h.ServeHTTP(w, r)
	a.Equal(w.Code, 1)

	// 访问不允许的域名
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "http://sub.caixw.io/test", nil)
	h.ServeHTTP(w, r)
	a.NotNil(w).NotNil(r)
	a.Equal(w.Code, http.StatusNotFound)

	// 访问不允许的域名
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "http://sub.1example.com/test", nil)
	h.ServeHTTP(w, r)
	a.NotNil(w).NotNil(r)
	a.Equal(w.Code, http.StatusNotFound)
}

func TestNew_empty_domains(t *testing.T) {
	a := assert.New(t)

	h := New(h1)
	a.NotNil(h)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "http://caixw.io/test", nil)
	a.NotNil(w).NotNil(r)
	h.ServeHTTP(w, r)
	a.Equal(w.Code, http.StatusNotFound)

	// 访问不允许的域名
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "http://not.exsits/test", nil)
	h.ServeHTTP(w, r)
	a.NotNil(w).NotNil(r)
	a.Equal(w.Code, http.StatusNotFound)
}
