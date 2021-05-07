// SPDX-License-Identifier: MIT

package errorhandler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

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
