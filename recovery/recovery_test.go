// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package recovery

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

var _ RecoverFunc = defaultRecoverFunc
var _ RecoverFunc = PrintDebug

var f1 = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(1)
}

var h1 = http.HandlerFunc(f1)

func TestDefaultRecoverFunc(t *testing.T) {
	a := assert.New(t)
	w := httptest.NewRecorder()
	a.NotNil(w)

	defaultRecoverFunc(w, "not found")
	a.Equal(http.StatusText(http.StatusInternalServerError)+"\n", w.Body.String())
}

func TestNew(t *testing.T) {
	a := assert.New(t)

	// h参数传递空值
	a.Panic(func() {
		New(nil, nil)
	})

	// 指定 fun 参数为 nil值，可以正常使用
	h := New(h1, nil)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "http://caixw.io/test", nil)
	a.NotNil(h).NotNil(w).NotNil(r)
	h.ServeHTTP(w, r)
	a.Equal(w.Code, 1)

	// 触发 panic
	h = New(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test")
	}), nil)
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "http://caixw.io/test", nil)
	a.NotNil(h).NotNil(w).NotNil(r)
	h.ServeHTTP(w, r)
	a.Equal(w.Code, http.StatusInternalServerError)
}

func TestPrintDebug(t *testing.T) {
	w := httptest.NewRecorder()
	PrintDebug(w, "PrintDebug")
	t.Log(w.Body.String())
}
