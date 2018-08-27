// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package compress

import (
	"compress/flate"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/issue9/assert"
)

var (
	_ WriterFunc = NewDeflate
	_ WriterFunc = NewGzip
)

var f1 = func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("f1\nf2"))
}

func TestCompress(t *testing.T) {
	a := assert.New(t)
	mgr := NewManager(map[string]WriterFunc{
		"gzip":    NewGzip,
		"deflate": NewDeflate,
	}, []string{"text"}, 0)
	srv := mgr.New(http.HandlerFunc(f1), log.New(os.Stderr, "", log.LstdFlags))
	a.NotNil(srv)

	// 未指定 accept-encoding
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	srv.ServeHTTP(w, r)
	a.Equal(w.Body.String(), "f1\nf2")
	a.Equal(w.Header().Get("Content-Encoding"), "")

	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Accept-encoding", "gzip;q=0.8,deflate")
	srv.ServeHTTP(w, r)
	a.Equal(w.Header().Get("Content-Encoding"), "deflate")
	a.NotEqual(w.Body.String(), "f1\nf2")
	// 解码后相等
	data, err := ioutil.ReadAll(flate.NewReader(w.Body))
	a.NotError(err).NotNil(data)
	a.Equal(string(data), "f1\nf2")
}
