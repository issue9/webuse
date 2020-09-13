// SPDX-License-Identifier: MIT

package host

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

var _ http.Handler = &Host{}

var f1 = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
}

var h1 = http.HandlerFunc(f1)

func TestNew(t *testing.T) {
	a := assert.New(t)

	h := New(h1, "caixw.io", "caixw.oi", "*.example.com")
	a.NotNil(h)
	a.Equal(len(h.domains), 2).
		Equal(len(h.wildcards), 1)

	// HTTP
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "http://caixw.io/test", nil)
	a.NotNil(w).NotNil(r)
	h.ServeHTTP(w, r)
	a.Equal(w.Code, http.StatusAccepted)

	// HTTPS
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "https://caixw.io/test", nil)
	a.NotNil(w).NotNil(r)
	h.ServeHTTP(w, r)
	a.Equal(w.Code, http.StatusAccepted)

	// 泛域名
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "https://xx.example.com/test", nil)
	a.NotNil(w).NotNil(r)
	h.ServeHTTP(w, r)
	a.Equal(w.Code, http.StatusAccepted)

	// 带端口
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "http://caixw.io:88/test", nil)
	a.NotNil(w).NotNil(r)
	h.ServeHTTP(w, r)
	a.Equal(w.Code, http.StatusAccepted)

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

func TestHost_add_delete(t *testing.T) {
	a := assert.New(t)

	h := New(h1)
	h.Add("xx.example.com")
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

	h := New(h1)
	a.NotNil(h)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "http://caixw.io/test", nil)
	a.NotNil(w).NotNil(r)
	h.ServeHTTP(w, r)
	a.Equal(w.Code, http.StatusNotFound)

	h.Omitempty(true)
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "http://caixw.io/test", nil)
	h.ServeHTTP(w, r)
	a.Equal(w.Code, http.StatusAccepted)

	// 访问不允许的域名
	h.Omitempty(false)
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "http://not.exsits/test", nil)
	h.ServeHTTP(w, r)
	a.NotNil(w).NotNil(r)
	a.Equal(w.Code, http.StatusNotFound)

	h.Omitempty(true)
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "http://not.exsits/test", nil)
	h.ServeHTTP(w, r)
	a.NotNil(w).NotNil(r)
	a.Equal(w.Code, http.StatusAccepted)
}
