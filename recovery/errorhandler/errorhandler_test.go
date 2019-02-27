// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package errorhandler

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"

	"github.com/issue9/middleware/recovery"
)

func testRenderError(w http.ResponseWriter, status int) {
	w.Header().Set("Content-type", "test")
	w.WriteHeader(status)
	w.Write([]byte("test"))
}

func TestErrorHandler_Add(t *testing.T) {
	a := assert.New(t)
	eh := New()

	a.NotError(eh.Add(nil, 500, 501))
	a.Error(eh.Add(nil, 500, 502)) // 已经存在

	a.NotError(eh.Add(testRenderError, 400, 401))
	a.Error(eh.Add(testRenderError, 401, 402)) // 已经存在
}

func TestErrorHandler_Set(t *testing.T) {
	a := assert.New(t)
	eh := New()

	eh.Set(nil, 500, 501)
	f, found := eh.handlers[500]
	a.True(found).Nil(f)

	eh.Set(testRenderError, 500, 502)
	a.Equal(eh.handlers[500], HandleFunc(testRenderError))
}

func TestErrorHandler_New(t *testing.T) {
	a := assert.New(t)
	eh := New()
	a.NotNil(eh)
	a.NotError(eh.Add(testRenderError, http.StatusBadRequest, http.StatusNotFound))

	f1 := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("f1"))
	}
	h1 := http.HandlerFunc(f1)

	// New，400 错误，不会采用 f1 的内容
	h := recovery.New(eh.New(h1), eh.Recovery(log.New(os.Stdout, "--", 0)))
	a.NotPanic(func() {
		srv := rest.NewServer(t, h, nil)
		srv.Get("/path").
			Do().
			Status(http.StatusBadRequest).
			Header("Content-Type", "test").
			StringBody("test")
	})

	// New，正常访问，采用 h 的内容
	h = recovery.New(eh.NewFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "h1")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("h"))
	}), eh.Recovery(nil))
	a.NotPanic(func() {
		srv := rest.NewServer(t, h, nil)
		srv.Get("/path").
			Do().
			Status(http.StatusOK).
			Header("Content-Type", "h1").
			StringBody("h")
	})

	// NewFunc，400 错误，不会采用 f1 的内容
	h = recovery.New(eh.NewFunc(f1), eh.Recovery(nil))
	a.NotPanic(func() {
		srv := rest.NewServer(t, h, nil)
		srv.Get("/path").
			Do().
			Status(http.StatusBadRequest).
			Header("Content-Type", "test").
			StringBody("test")
	})
}

func TestErrorHandler_Render(t *testing.T) {
	a := assert.New(t)
	eh := New()

	w := httptest.NewRecorder()
	eh.Render(w, http.StatusOK)
	a.Equal(w.Code, http.StatusOK)

	w = httptest.NewRecorder()
	eh.Render(w, http.StatusInternalServerError)
	a.Equal(w.Code, http.StatusInternalServerError)

	// 设置为空，依然采用 defaultRender
	eh.Set(nil, http.StatusInternalServerError)
	w = httptest.NewRecorder()
	eh.Render(w, http.StatusInternalServerError)
	a.Equal(w.Code, http.StatusInternalServerError)

	// 设置为 testRenderError
	eh.Set(testRenderError, http.StatusInternalServerError)
	w = httptest.NewRecorder()
	eh.Render(w, http.StatusInternalServerError)
	a.Equal(w.Code, http.StatusInternalServerError).
		Equal(w.Header().Get("Content-Type"), "test").
		Equal(w.Body.String(), "test")
}

func TestErrorHandler_Render_0(t *testing.T) {
	a := assert.New(t)
	eh := New()

	eh.Add(testRenderError, 401, 402)
	w := httptest.NewRecorder()
	eh.Render(w, 401)
	a.Equal(w.Code, 401).
		Equal(w.Header().Get("Content-Type"), "test").
		Equal(w.Body.String(), "test")
	w = httptest.NewRecorder()
	eh.Render(w, 405) // 不存在
	a.Equal(w.Code, 405)

	// 设置为 testRender
	eh.Set(testRenderError, 0, 401, 402)
	w = httptest.NewRecorder()
	eh.Render(w, 401)
	a.Equal(w.Code, 401).
		Equal(w.Header().Get("Content-Type"), "test").
		Equal(w.Body.String(), "test")
	w = httptest.NewRecorder()
	eh.Render(w, 405) // 采用 0
	a.Equal(w.Code, 405).
		Equal(w.Header().Get("Content-Type"), "test").
		Equal(w.Body.String(), "test")
}

func TestErrorHandler_Recovery(t *testing.T) {
	a := assert.New(t)
	eh := New()

	fn := eh.Recovery(nil)

	// 普通内容
	w := httptest.NewRecorder()
	a.NotPanic(func() { fn(w, "msg") })
	a.Equal(w.Result().StatusCode, http.StatusInternalServerError)

	// 普通数值
	w = httptest.NewRecorder()
	a.NotPanic(func() { fn(w, http.StatusBadGateway) })
	a.Equal(w.Result().StatusCode, http.StatusInternalServerError)

	// httpStatus
	w = httptest.NewRecorder()
	a.NotPanic(func() { fn(w, httpStatus(http.StatusBadGateway)) })
	a.Equal(w.Result().StatusCode, http.StatusBadGateway)

	// 以下为带 errlog 的测试内容

	errlog := new(bytes.Buffer)
	fn = eh.Recovery(log.New(errlog, "", 0))

	// 普通内容
	w = httptest.NewRecorder()
	a.NotPanic(func() { fn(w, "msg") })
	a.Equal(w.Result().StatusCode, http.StatusInternalServerError)
	a.True(strings.Contains(errlog.String(), "msg"))

	// 普通数值
	w = httptest.NewRecorder()
	errlog.Reset()
	a.NotPanic(func() { fn(w, http.StatusBadGateway) })
	a.Equal(w.Result().StatusCode, http.StatusInternalServerError)
	a.True(strings.Contains(errlog.String(), strconv.FormatInt(http.StatusBadGateway, 10)))

	// httpStatus，没有输出日志，算是正常退出。
	w = httptest.NewRecorder()
	errlog.Reset()
	a.NotPanic(func() { fn(w, httpStatus(http.StatusBadGateway)) })
	a.Equal(w.Result().StatusCode, http.StatusBadGateway)
	a.Empty(errlog.String())
}
