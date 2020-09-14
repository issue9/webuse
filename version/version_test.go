// SPDX-License-Identifier: MIT

package version

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/middleware"
)

var _ middleware.Middlewarer = &Version{}

var f1 = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(1)
}

func BenchmarkFindVersionNumber(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = findVersionNumber("application/json;version=1.0;application/json")
	}
}

func TestVersion(t *testing.T) {
	a := assert.New(t)

	h := &Version{Version: "1.0", Strict: true}

	// 相同版本号
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "http://caixw.io/test", nil)
	r.Header.Set("Accept", "application/json; version=1.0")
	a.NotNil(w).NotNil(r)
	h.MiddlewareFunc(f1).ServeHTTP(w, r)
	a.Equal(w.Code, 1)

	// 空版本
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "http://not.exsits/test", nil)
	r.Header.Set("Accept", "application/json; version=")
	h.MiddlewareFunc(f1).ServeHTTP(w, r)
	a.NotNil(w).NotNil(r)
	a.Equal(w.Code, http.StatusForbidden)

	// 不同版本
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "http://not.exsits/test", nil)
	r.Header.Set("Accept", "application/json; version=2")
	h.MiddlewareFunc(f1).ServeHTTP(w, r)
	a.NotNil(w).NotNil(r)
	a.Equal(w.Code, http.StatusNotFound)

	// strict == false

	h.Strict = false

	// 相同版本号
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "http://caixw.io/test", nil)
	r.Header.Set("Accept", "application/json; version=1.0")
	a.NotNil(w).NotNil(r)
	h.MiddlewareFunc(f1).ServeHTTP(w, r)
	a.Equal(w.Code, 1)

	// 空版本
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "http://not.exsits/test", nil)
	r.Header.Set("Accept", "application/json; version=")
	h.MiddlewareFunc(f1).ServeHTTP(w, r)
	a.NotNil(w).NotNil(r)
	a.Equal(w.Code, 1)

	// 不同版本
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "http://not.exsits/test", nil)
	r.Header.Set("Accept", "application/json; version=2")
	h.MiddlewareFunc(f1).ServeHTTP(w, r)
	a.NotNil(w).NotNil(r)
	a.Equal(w.Code, http.StatusNotFound)
}

func TestFindVersionNumber(t *testing.T) {
	a := assert.New(t)

	a.Equal(findVersionNumber(""), "")
	a.Equal(findVersionNumber("version="), "")
	a.Equal(findVersionNumber("Version="), "")
	a.Equal(findVersionNumber(";version="), "")
	a.Equal(findVersionNumber(";version=;"), "")
	a.Equal(findVersionNumber(";version=1.0"), "1.0")
	a.Equal(findVersionNumber(";version=1.0;"), "1.0")
	a.Equal(findVersionNumber(";version=1.0;application/json"), "1.0")
	a.Equal(findVersionNumber("application/json;version=1.0"), "1.0")
	a.Equal(findVersionNumber("application/json;version=1.0;application/json"), "1.0")
}
