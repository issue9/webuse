// SPDX-License-Identifier: MIT

package errorhandler

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"

	"github.com/issue9/middleware/v5"
)

func errorHandlerFunc(w io.Writer, status int) {
	w.Write([]byte("test"))
}

func TestErrorHandler_Add(t *testing.T) {
	a := assert.New(t, false)
	eh := New()

	a.Panic(func() {
		eh.Add(nil, 500, 501)
	})
	a.True(eh.Add(errorHandlerFunc, 500, 501))
	a.True(eh.Exists(500)).True(eh.Exists(501))
	a.False(eh.Add(errorHandlerFunc, 500, 502)) // 已经存在
	a.True(eh.Exists(500)).True(eh.Exists(501))

	a.True(eh.Add(errorHandlerFunc, 400, 401))
	a.False(eh.Add(errorHandlerFunc, 401, 402)) // 已经存在
}

func TestErrorHandler_Set(t *testing.T) {
	a := assert.New(t, false)
	eh := New()

	eh.Set(nil, 500, 501)
	a.False(eh.Exists(500)).False(eh.Exists(501))

	eh.Set(errorHandlerFunc, 500, 502)
	a.True(eh.Exists(500)).False(eh.Exists(501))
	a.Equal(eh.handlers[500], HandleFunc(errorHandlerFunc))

	eh.Set(nil, 500, 501)
	a.False(eh.Exists(500)).True(eh.Exists(502))
}

func TestErrorHandler_MiddlewareFunc(t *testing.T) {
	a := assert.New(t, false)
	eh := New()
	a.NotNil(eh)
	a.True(eh.Add(errorHandlerFunc, http.StatusBadRequest, http.StatusNotFound, http.StatusAccepted))

	// 400 错误，不会采用 f400 的内容，而是 errorHandlerFunc
	ms := middleware.NewMiddlewares(rest.BuildHandler(a, http.StatusBadRequest, "400", nil))
	ms.Append(eh.Renderer()).Append(eh.Middleware)
	srv := rest.NewServer(a, ms, nil)
	srv.Get("/path").
		Do(nil).
		Status(http.StatusBadRequest).
		StringBody("test")

	// 202 错误，不会采用 f202 的内容，而是 errorHandlerFunc
	ms = middleware.NewMiddlewares(rest.BuildHandler(a, http.StatusAccepted, "202", nil))
	ms.Append(eh.Renderer()).Append(eh.Middleware)
	srv = rest.NewServer(a, ms, nil)
	srv.Get("/path").
		Do(nil).
		Status(http.StatusAccepted).
		StringBody("test")

	// 正常访问，采用 h 的内容
	ms = middleware.NewMiddlewares(rest.BuildHandler(a, http.StatusOK, "200", map[string]string{"Server": "h1"}))
	ms.Append(eh.Renderer()).Append(eh.Middleware)
	srv = rest.NewServer(a, ms, nil)
	srv.Get("/path").
		Do(nil).
		Status(http.StatusOK).
		Header("Server", "h1").
		StringBody("200")

	// 正常访问，采用 h 的内容
	ms = middleware.NewMiddlewares(rest.BuildHandler(a, http.StatusNotAcceptable, "204", map[string]string{"Server": "h1"}))
	ms.Append(eh.Renderer()).Append(eh.Middleware)
	srv = rest.NewServer(a, ms, nil)
	srv.Get("/path").
		Do(nil).
		Status(http.StatusNotAcceptable).
		Header("Server", "h1")
}

func TestErrorHandler_Render(t *testing.T) {
	a := assert.New(t, false)
	eh := New()

	w := &bytes.Buffer{}
	eh.Render(w, http.StatusOK)
	a.Equal(w.String(), http.StatusText(http.StatusOK))

	w.Reset()
	eh.Render(w, http.StatusInternalServerError)
	a.Equal(w.String(), http.StatusText(http.StatusInternalServerError))

	// 设置为空，依然采用 defaultRender
	w.Reset()
	eh.Set(nil, http.StatusInternalServerError)
	eh.Render(w, http.StatusInternalServerError)
	a.Equal(w.String(), http.StatusText(http.StatusInternalServerError))

	// 设置为 errorHandlerFunc
	w.Reset()
	eh.Set(errorHandlerFunc, http.StatusInternalServerError)
	eh.Render(w, http.StatusInternalServerError)
	a.Equal(w.String(), "test")
}
