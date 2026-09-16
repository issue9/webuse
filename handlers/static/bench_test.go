// SPDX-FileCopyrightText: 2014-2026 caixw
//
// SPDX-License-Identifier: MIT

package static

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/issue9/assert/v5"
	"github.com/issue9/mux/v9/types"

	"github.com/issue9/webuse/v7/internal/testserver"
)

func BenchmarkServeFile(b *testing.B) {
	a := assert.New(b, false)
	s := testserver.New(a)

	fsys := os.DirFS("./testdata")

	b.ReportAllocs()

	h := ServeFileHandler(fsys, "name", "default.html")

	for b.Loop() {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "index.html", nil)
		h(s.NewContext(w, r, types.NewContext()))
	}
}
