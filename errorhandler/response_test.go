// SPDX-License-Identifier: MIT

package errorhandler

import (
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
