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

var f1 = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(1)
}

var h1 = http.HandlerFunc(f1)

func TestHostFunc(t *testing.T) {
	a := assert.New(t)

	h := HostFunc(f1, "caixw.io", "caixw.oi")
	a.NotNil(h)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "http://caixw.io/test", nil)
	a.NotNil(w).NotNil(r)
	h.ServeHTTP(w, r)
	a.Equal(w.Code, 1)

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "https://caixw.io/test", nil)
	a.NotNil(w).NotNil(r)
	h.ServeHTTP(w, r)
	a.Equal(w.Code, 1)

	// 访问不允许的域名
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "http://not.exsits/test", nil)
	h.ServeHTTP(w, r)
	a.NotNil(w).NotNil(r)
	a.Equal(w.Code, http.StatusForbidden)
}

func TestHostFunc_empty_domains(t *testing.T) {
	a := assert.New(t)

	h := HostFunc(f1)
	a.NotNil(h)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "http://caixw.io/test", nil)
	a.NotNil(w).NotNil(r)
	h.ServeHTTP(w, r)
	a.Equal(w.Code, http.StatusForbidden)

	// 访问不允许的域名
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "http://not.exsits/test", nil)
	h.ServeHTTP(w, r)
	a.NotNil(w).NotNil(r)
	a.Equal(w.Code, http.StatusForbidden)
}
