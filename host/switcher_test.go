// SPDX-License-Identifier: MIT

package host

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

var f2 = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(2)
}

var h2 = http.HandlerFunc(f2)

func TestNewSwitcher(t *testing.T) {
	a := assert.New(t)

	switcher := NewSwitcher()
	a.NotNil(switcher)

	switcher.AddHost(h1, "caixw.io", "*.example.com")
	switcher.AddHost(h2, "caixw.oi")

	// h1
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "https://caixw.io/test", nil)
	a.NotNil(w).NotNil(r)
	switcher.ServeHTTP(w, r)
	a.Equal(w.Code, 1)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "https://sub.example.com/test", nil)
	a.NotNil(w).NotNil(r)
	switcher.ServeHTTP(w, r)
	a.Equal(w.Code, 1)

	// h2
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "http://caixw.oi/test", nil)
	a.NotNil(w).NotNil(r)
	switcher.ServeHTTP(w, r)
	a.Equal(w.Code, 2)

	// 未带域名
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "http://not.exsits/test", nil)
	switcher.ServeHTTP(w, r)
	a.NotNil(w).NotNil(r)
	a.Equal(w.Code, http.StatusNotFound)
}
