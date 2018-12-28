// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package basic

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/middleware/auth"
)

var _ http.Handler = &basic{}

var (
	fok = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}

	hok = http.HandlerFunc(fok)

	authFunc = func(username, passowrd []byte) (interface{}, bool) {
		return username, true
	}
)

func TestNew(t *testing.T) {
	a := assert.New(t)
	var h http.Handler

	a.Panic(func() {
		h = New(nil, nil, "", false, nil)
	})

	a.Panic(func() {
		h = New(hok, nil, "", false, nil)
	})

	a.NotPanic(func() {
		h = New(hok, authFunc, "", false, nil)
	})

	bb, ok := h.(*basic)
	a.True(ok).
		Equal(bb.authorization, "Authorization").
		Equal(bb.authenticate, "WWW-Authenticate").
		Equal(bb.unauthorizationStatus, http.StatusUnauthorized).
		Nil(bb.errlog).
		NotNil(bb.auth)

	a.NotPanic(func() {
		h = New(hok, authFunc, "", true, log.New(ioutil.Discard, "", 0))
	})

	bb, ok = h.(*basic)
	a.True(ok).
		Equal(bb.authorization, "Proxy-Authorization").
		Equal(bb.authenticate, "Proxy-Authenticate").
		Equal(bb.unauthorizationStatus, http.StatusProxyAuthRequired).
		NotNil(bb.errlog).
		NotNil(bb.auth)
}

func TestServeHTTP_ok(t *testing.T) {
	a := assert.New(t)
	var h http.Handler

	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := r.Context().Value(auth.ValueKey).([]byte)
		a.Equal(string(username), "Aladdin")
	})

	a.NotPanic(func() {
		h = New(ok, authFunc, "example.com", false, nil)
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	h.ServeHTTP(w, r)
	a.Equal(w.Header().Get("WWW-Authenticate"), `Basic realm="example.com"`).
		Equal(http.StatusUnauthorized, w.Code)

	// 正确的访问
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	// Aladdin, open sesame，来自 https://zh.wikipedia.org/wiki/HTTP基本认证
	r.Header.Set("Authorization", "Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==")
	h.ServeHTTP(w, r)
}

func TestServeHTTP_faild(t *testing.T) {
	a := assert.New(t)
	var h http.Handler

	faild := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		obj := r.Context().Value(auth.ValueKey)
		a.Nil(obj)
	})

	a.NotPanic(func() {
		h = New(faild, authFunc, "example.com", false, nil)
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	h.ServeHTTP(w, r)
	a.Equal(w.Header().Get("WWW-Authenticate"), `Basic realm="example.com"`).
		Equal(http.StatusUnauthorized, w.Code)

	// 错误的编码
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Authorization", "Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ===")
	h.ServeHTTP(w, r)
}
