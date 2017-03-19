// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

func TestVersionFunc_strict(t *testing.T) {
	a := assert.New(t)

	h := VersionFunc(f1, "1.0", true)
	a.NotNil(h)

	// 相同版本号
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "http://caixw.io/test", nil)
	r.Header.Set("Accept", "application/json; version=1.0")
	a.NotNil(w).NotNil(r)
	h.ServeHTTP(w, r)
	a.Equal(w.Code, 1)

	// 空版本
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "http://not.exsits/test", nil)
	r.Header.Set("Accept", "application/json; version=")
	h.ServeHTTP(w, r)
	a.NotNil(w).NotNil(r)
	a.Equal(w.Code, http.StatusForbidden)

	// 不同版本
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "http://not.exsits/test", nil)
	r.Header.Set("Accept", "application/json; version=2")
	h.ServeHTTP(w, r)
	a.NotNil(w).NotNil(r)
	a.Equal(w.Code, http.StatusForbidden)
}

func TestVersionFunc_nostrict(t *testing.T) {
	a := assert.New(t)

	h := VersionFunc(f1, "1.0", false)
	a.NotNil(h)

	// 相同版本号
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "http://caixw.io/test", nil)
	r.Header.Set("Accept", "application/json; version=1.0")
	a.NotNil(w).NotNil(r)
	h.ServeHTTP(w, r)
	a.Equal(w.Code, 1)

	// 空版本
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "http://not.exsits/test", nil)
	r.Header.Set("Accept", "application/json; version=")
	h.ServeHTTP(w, r)
	a.NotNil(w).NotNil(r)
	a.Equal(w.Code, 1)

	// 不同版本
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "http://not.exsits/test", nil)
	r.Header.Set("Accept", "application/json; version=2")
	h.ServeHTTP(w, r)
	a.NotNil(w).NotNil(r)
	a.Equal(w.Code, http.StatusForbidden)
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
