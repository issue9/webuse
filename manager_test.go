// SPDX-License-Identifier: MIT

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

func buildMiddleware(a *assert.Assertion, text string) Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			a.NotError(w.Write([]byte(text)))
			h.ServeHTTP(w, r)
		})
	}
}

func buildHandler(a *assert.Assertion, code int, content string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
		a.NotError(w.Write([]byte(content)))
	})
}

func TestManager(t *testing.T) {
	a := assert.New(t)

	m := NewManager(buildHandler(a, http.StatusCreated, "test"))
	a.NotNil(m)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Code, http.StatusCreated). // 没有中间件
						Equal(w.Body.String(), "test")

	w = httptest.NewRecorder()
	m.After(buildMiddleware(a, "a0"))
	m.ServeHTTP(w, r)
	a.Equal(w.Code, http.StatusOK). // 中间件有输出，将状态码改为 200
					Equal(w.Body.String(), "a0test")

	// 执行过程中添加中间件
	m.After(buildMiddleware(a, "a1"))
	m.After(buildMiddleware(a, "a2"))
	m.Before(buildMiddleware(a, "b1"))
	m.Before(buildMiddleware(a, "b2"))
	w = httptest.NewRecorder()
	m.ServeHTTP(w, r)
	a.Equal(w.Code, http.StatusOK). // 中间件有输出，将状态码改为 200
					Equal(w.Body.String(), "b2b1a0a1a2test")

	// 重置中间件。同时状态码输出也改为 1
	m.Reset()
	w = httptest.NewRecorder()
	m.ServeHTTP(w, r)
	a.Equal(w.Code, http.StatusCreated).
		Equal(w.Body.String(), "test")

	// 执行过程中添加中间件
	m.After(buildMiddleware(a, "m2"))
	w = httptest.NewRecorder()
	m.ServeHTTP(w, r)
	a.Equal(w.Code, http.StatusOK). // 中间件有输出，将状态码改为 200
					Equal(w.Body.String(), "m2test")
}
