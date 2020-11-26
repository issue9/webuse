// SPDX-License-Identifier: MIT

package basic

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"

	"github.com/issue9/middleware/v3/auth"
)

var (
	authFunc = func(username, password []byte) (interface{}, bool) {
		return username, true
	}
)

func TestNew(t *testing.T) {
	a := assert.New(t)
	var b *Basic

	a.Panic(func() {
		b = New(nil, "", false, nil)
	})

	a.NotPanic(func() {
		b = New(authFunc, "", false, nil)
	})

	a.Equal(b.authorization, "Authorization").
		Equal(b.authenticate, "WWW-Authenticate").
		Equal(b.unauthorizationStatus, http.StatusUnauthorized).
		Nil(b.errlog).
		NotNil(b.auth)

	a.NotPanic(func() {
		b = New(authFunc, "", true, log.New(ioutil.Discard, "", 0))
	})

	a.Equal(b.authorization, "Proxy-Authorization").
		Equal(b.authenticate, "Proxy-Authenticate").
		Equal(b.unauthorizationStatus, http.StatusProxyAuthRequired).
		NotNil(b.errlog).
		NotNil(b.auth)
}

func TestServeHTTP_ok(t *testing.T) {
	a := assert.New(t)

	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := r.Context().Value(auth.ValueKey).([]byte)
		a.Equal(string(username), "Aladdin")
	})

	b := New(authFunc, "example.com", false, nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	b.Middleware(ok).ServeHTTP(w, r)
	a.Equal(w.Header().Get("WWW-Authenticate"), `Basic realm="example.com"`).
		Equal(http.StatusUnauthorized, w.Code)

	// 正确的访问
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	// Aladdin, open sesame，来自 https://zh.wikipedia.org/wiki/HTTP基本认证
	r.Header.Set("Authorization", "Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==")
	b.Middleware(ok).ServeHTTP(w, r)
}

func TestServeHTTP_failed(t *testing.T) {
	a := assert.New(t)

	failed := func(w http.ResponseWriter, r *http.Request) {
		obj := r.Context().Value(auth.ValueKey)
		a.Nil(obj)
	}

	b := New(authFunc, "example.com", false, nil)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	b.MiddlewareFunc(failed).ServeHTTP(w, r)
	a.Equal(w.Header().Get("WWW-Authenticate"), `Basic realm="example.com"`).
		Equal(http.StatusUnauthorized, w.Code)

	// 错误的编码
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Authorization", "Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ===")
	b.MiddlewareFunc(failed).ServeHTTP(w, r)
}
