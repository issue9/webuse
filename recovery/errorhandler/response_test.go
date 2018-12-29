// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

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
