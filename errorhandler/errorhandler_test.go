// SPDX-License-Identifier: MIT

package errorhandler

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"

	"github.com/issue9/middleware/v5/recovery"
)

func errorHandlerFunc(w io.Writer, status int) {
	w.Write([]byte("test"))
}

func TestErrorHandler_Add(t *testing.T) {
	a := assert.New(t)
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
	a := assert.New(t)
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
	a := assert.New(t)
	eh := New()
	a.NotNil(eh)
	a.NotError(eh.Add(errorHandlerFunc, http.StatusBadRequest, http.StatusNotFound, http.StatusAccepted))

	f400 := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400"))
	}

	f202 := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("400"))
	}

	// MiddlewareFunc，400 错误，不会采用 f400 的内容，而是 errorHandlerFunc
	h := eh.Recovery(nil).Middleware(eh.MiddlewareFunc(f400))
	srv := rest.NewServer(t, h, nil)
	srv.Get("/path").
		Do().
		Status(http.StatusBadRequest).
		StringBody("test")

	// MiddlewareFunc，202 错误，不会采用 f202 的内容，而是 errorHandlerFunc
	h = eh.Recovery(nil).Middleware(eh.MiddlewareFunc(f202))
	srv = rest.NewServer(t, h, nil)
	srv.Get("/path").
		Do().
		Status(http.StatusAccepted).
		StringBody("test")

	// MiddlewareFunc，正常访问，采用 h 的内容
	h = eh.Recovery(nil).MiddlewareFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "h1")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("h"))
	})
	srv = rest.NewServer(t, h, nil)
	srv.Get("/path").
		Do().
		Status(http.StatusOK).
		Header("Content-Type", "h1").
		StringBody("h")

	// MiddlewareFunc，正常访问，采用 h 的内容
	h = eh.Recovery(nil).MiddlewareFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "h1")
		w.WriteHeader(http.StatusNoContent)
	})
	srv = rest.NewServer(t, h, nil)
	srv.Get("/path").
		Do().
		Status(http.StatusNoContent).
		Header("Content-Type", "h1")

	// recovery.DefaultRecover 并不会正常处理 errorhandler 的状态码错误
	// NOTE: recovery.DefaultRecover 的 WriteHeader 会与 errorhandler 中的相关突。
	h = recovery.DefaultRecover(http.StatusInternalServerError).Middleware(eh.MiddlewareFunc(f400))
	srv = rest.NewServer(t, h, nil)
	srv.Get("/path").
		Do().
		Status(http.StatusBadRequest).                                              // 报头已经在 errorhandler 中输出
		StringBody("test" + http.StatusText(http.StatusInternalServerError) + "\n") // 输出内容也是结合了 errorhandler 和 recovery
}

func TestErrorHandler_Render(t *testing.T) {
	a := assert.New(t)
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
	a.Equal(w.Result().StatusCode, http.StatusOK) // 不需要调用 recovery，输出黑认的 200

	// 以下为自定义 rf 参数

	fn = eh.Recovery(recovery.ConsoleRecover(http.StatusNotFound))

	// 普通内容
	w = httptest.NewRecorder()
	a.NotPanic(func() { fn(w, "msg") })
	a.Equal(w.Result().StatusCode, http.StatusNotFound)
	msg, err := ioutil.ReadAll(w.Result().Body)
	a.NotError(err).NotNil(msg)
	a.True(strings.Contains(string(msg), "msg"))

	// 普通数值
	w = httptest.NewRecorder()
	a.NotPanic(func() { fn(w, http.StatusBadGateway) })
	a.Equal(w.Result().StatusCode, http.StatusNotFound)
	msg, err = ioutil.ReadAll(w.Result().Body)
	a.NotError(err).NotNil(msg)
	a.True(strings.Contains(string(msg), strconv.FormatInt(http.StatusBadGateway, 10)))

	// httpStatus，没有输出日志，算是正常退出。
	w = httptest.NewRecorder()
	a.NotPanic(func() { fn(w, httpStatus(http.StatusBadGateway)) })
	a.Equal(w.Result().StatusCode, http.StatusOK)
	msg, err = ioutil.ReadAll(w.Result().Body)
	a.NotError(err).Empty(msg)
}
