// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package compress

import (
	"bytes"
	"compress/flate"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"
)

var (
	_ WriterFunc = NewDeflate
	_ WriterFunc = NewGzip
)

var f1 = func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte("f1\nf2"))
}

func TestCompress(t *testing.T) {
	a := assert.New(t)
	opt := &Options{
		Funcs: map[string]WriterFunc{
			"gzip":    NewGzip,
			"deflate": NewDeflate,
		},
		Types:    []string{"text/*"},
		Size:     0,
		ErrorLog: log.New(os.Stderr, "", log.LstdFlags),
	}
	srv := rest.NewServer(t, New(http.HandlerFunc(f1), opt), nil)

	// 指定 accept-encoding = *
	srv.NewRequest(http.MethodGet, "/").
		Header("Content-Type", "text/html").
		Header("Accept-Encoding", "*").
		Do().
		StringBody("f1\nf2").
		Header("Content-Encoding", "")

	// 指定 accept-encoding = identity
	srv.NewRequest(http.MethodGet, "/").
		Header("Content-Type", "text/html").
		Header("Accept-Encoding", "identity").
		Do().
		StringBody("f1\nf2").
		Header("Content-Encoding", "")

	// 指定 accept-encoding 为空
	srv.NewRequest(http.MethodGet, "/").
		Header("Content-Type", "text/html").
		Header("Accept-Encoding", "").
		Do().
		StringBody("f1\nf2").
		Header("Content-Encoding", "")

	// accept-encoding = deflate
	buf := new(bytes.Buffer)
	srv.NewRequest(http.MethodGet, "/").
		Header("Content-Type", "text/html").
		Header("Accept-encoding", "gzip;q=0.8,deflate").
		Do().
		BodyNotNil().
		ReadBody(buf).
		Header("Content-Encoding", "deflate").
		Header("Content-Type", "text/html").
		Header("Vary", "Content-Encoding")

	// 解码后相等
	a.True(len(buf.Bytes()) > 0)
	data, err := ioutil.ReadAll(flate.NewReader(buf))
	a.NotError(err).NotNil(data)
	a.Equal(string(data), "f1\nf2")
}
