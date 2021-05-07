// SPDX-License-Identifier: MIT

package errorhandler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

func TestResponse_WriteHeader(t *testing.T) {
	a := assert.New(t)

	eh := New()
	a.NotNil(eh)
	a.True(eh.Add(errorHandlerFunc, http.StatusInternalServerError))

	f201 := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, err := w.Write([]byte("201"))
		a.NotError(err)
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	eh.Middleware(http.HandlerFunc(f201)).ServeHTTP(w, r)
	a.Equal(w.Body.String(), "201").Equal(w.Code, http.StatusCreated)

	f500 := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("500"))
		a.NotError(err)
	}
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	a.PanicType(func() {
		eh.Middleware(http.HandlerFunc(f500)).ServeHTTP(w, r)
	}, httpStatus(http.StatusInternalServerError))
}

func TestWriteHeader(t *testing.T) {
	a := assert.New(t)

	w := httptest.NewRecorder()
	resp := &response{ResponseWriter: w}
	WriteHeader(resp, 200)
	a.Equal(w.Code, 200)

	w = httptest.NewRecorder()
	resp = &response{ResponseWriter: w}
	WriteHeader(resp, 400)
	a.Equal(w.Code, 400)

	w = httptest.NewRecorder()
	WriteHeader(w, 400)
	a.Equal(w.Code, 400)
}

func TestErrorHandler_Exit(t *testing.T) {
	a := assert.New(t)

	eh := New()
	w := httptest.NewRecorder()

	a.Panic(func() { eh.Exit(w, 5) })
	a.Panic(func() { eh.Exit(w, 0) })

	func() {
		defer func() {
			msg := recover()
			val, ok := msg.(httpStatus)
			a.True(ok).Equal(val, 0)
		}()

		eh.Exit(w, http.StatusNotFound)
	}()
}
