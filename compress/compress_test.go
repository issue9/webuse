// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package compress

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
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
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("f1\nf2"))
}

var f2 = func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte("f1\nf2"))
}

var f3 = func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("f1\nf2"))
}

var f4 = func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("f1\nf2"))
}

func TestCompress_f1(t *testing.T) {
	a := assert.New(t)
	opt := &Options{
		Funcs: map[string]WriterFunc{
			"gzip":    NewGzip,
			"deflate": NewDeflate,
		},
		Types:    []string{"text/*"},
		ErrorLog: log.New(os.Stderr, "", log.LstdFlags),
	}
	srv := rest.NewServer(t, New(http.HandlerFunc(f1), opt), nil)

	// 指定 accept-encoding = *
	srv.NewRequest(http.MethodGet, "/").
		Header("Accept-Encoding", "*").
		Do().
		StringBody("f1\nf2").
		Status(http.StatusAccepted).
		Header("Content-Encoding", "")

	// 指定 accept-encoding = identity
	srv.NewRequest(http.MethodGet, "/").
		Header("Accept-Encoding", "identity").
		Do().
		StringBody("f1\nf2").
		Status(http.StatusAccepted).
		Header("Content-Encoding", "")

	// 指定 accept-encoding 为空
	srv.NewRequest(http.MethodGet, "/").
		Header("Accept-Encoding", "").
		Do().
		StringBody("f1\nf2").
		Status(http.StatusAccepted).
		Header("Content-Encoding", "")

	// accept-encoding = deflate
	buf := new(bytes.Buffer)
	srv.NewRequest(http.MethodGet, "/").
		Header("Accept-encoding", "gzip;q=0.8,deflate").
		Do().
		BodyNotNil().
		ReadBody(buf).
		Status(http.StatusAccepted).
		Header("Content-Type", "text/html").
		Header("Content-Encoding", "deflate").
		Header("Vary", "Content-Encoding")

	// 解码后相等
	a.True(len(buf.Bytes()) > 0)
	data, err := ioutil.ReadAll(flate.NewReader(buf))
	a.NotError(err).NotNil(data)
	a.Equal(string(data), "f1\nf2")

	// accept-encoding = gzip
	buf = new(bytes.Buffer)
	srv.NewRequest(http.MethodGet, "/").
		Header("Accept-encoding", "gzip,deflate;q=0.8").
		Do().
		BodyNotNil().
		ReadBody(buf).
		Status(http.StatusAccepted).
		Header("Content-Type", "text/html").
		Header("Content-Encoding", "gzip").
		Header("Vary", "Content-Encoding")

	// 解码后相等
	a.True(len(buf.Bytes()) > 0)
	reader, err := gzip.NewReader(buf)
	a.NotError(err).NotNil(reader)
	data, err = ioutil.ReadAll(reader)
	a.NotError(err).NotNil(data)
	a.Equal(string(data), "f1\nf2")
}

func TestCompress_f2(t *testing.T) {
	a := assert.New(t)
	opt := &Options{
		Funcs: map[string]WriterFunc{
			"gzip":    NewGzip,
			"deflate": NewDeflate,
		},
		Types:    []string{"text/*"},
		ErrorLog: log.New(os.Stderr, "", log.LstdFlags),
	}
	srv := rest.NewServer(t, New(http.HandlerFunc(f2), opt), nil)

	// 指定 accept-encoding = *
	srv.NewRequest(http.MethodGet, "/").
		Header("Accept-Encoding", "*").
		Do().
		StringBody("f1\nf2").
		Status(http.StatusOK).
		Header("Content-Encoding", "")

	// 指定 accept-encoding = identity
	srv.NewRequest(http.MethodGet, "/").
		Header("Accept-Encoding", "identity").
		Do().
		StringBody("f1\nf2").
		Status(http.StatusOK).
		Header("Content-Encoding", "")

	// 指定 accept-encoding 为空
	srv.NewRequest(http.MethodGet, "/").
		Header("Accept-Encoding", "").
		Do().
		Status(http.StatusOK).
		StringBody("f1\nf2").
		Header("Content-Encoding", "")

	// accept-encoding = deflate
	buf := new(bytes.Buffer)
	srv.NewRequest(http.MethodGet, "/").
		Header("Accept-encoding", "gzip,deflate;q=0.8").
		Do().
		Status(http.StatusOK).
		BodyNotNil().
		ReadBody(buf).
		Header("Content-Type", "text/html").
		Header("Content-Encoding", "gzip").
		Header("Vary", "Content-Encoding")

	// 解码后相等
	a.True(len(buf.Bytes()) > 0)
	reader, err := gzip.NewReader(buf)
	a.NotError(err).NotNil(reader)
	data, err := ioutil.ReadAll(reader)
	a.NotError(err).NotNil(data)
	a.Equal(string(data), "f1\nf2")
}

func TestCompress_f3(t *testing.T) {
	a := assert.New(t)
	opt := &Options{
		Funcs: map[string]WriterFunc{
			"gzip":    NewGzip,
			"deflate": NewDeflate,
		},
		Types:    []string{"text/*"},
		ErrorLog: log.New(os.Stderr, "", log.LstdFlags),
	}
	srv := rest.NewServer(t, New(http.HandlerFunc(f3), opt), nil)

	// 指定 accept-encoding = *
	srv.NewRequest(http.MethodGet, "/").
		Header("Accept-Encoding", "*").
		Do().
		Status(http.StatusOK).
		StringBody("f1\nf2").
		Header("Content-Encoding", "")

	// 指定 accept-encoding = identity
	srv.NewRequest(http.MethodGet, "/").
		Header("Accept-Encoding", "identity").
		Do().
		Status(http.StatusOK).
		StringBody("f1\nf2").
		Header("Content-Encoding", "")

	// 指定 accept-encoding 为空
	srv.NewRequest(http.MethodGet, "/").
		Header("Accept-Encoding", "").
		Do().
		Status(http.StatusOK).
		StringBody("f1\nf2").
		Header("Content-Encoding", "")

	// accept-encoding = deflate
	buf := new(bytes.Buffer)
	srv.NewRequest(http.MethodGet, "/").
		Header("Accept-encoding", "gzip;q=0.8,deflate").
		Do().
		Status(http.StatusOK).
		BodyNotNil().
		ReadBody(buf).
		Header("Content-Type", "text/plain; charset=utf-8").
		Header("Content-Encoding", "deflate").
		Header("Vary", "Content-Encoding")

	// 解码后相等
	a.True(len(buf.Bytes()) > 0)
	data, err := ioutil.ReadAll(flate.NewReader(buf))
	a.NotError(err).NotNil(data)
	a.Equal(string(data), "f1\nf2")
}

func TestCompress_f4(t *testing.T) {
	opt := &Options{
		Funcs: map[string]WriterFunc{
			"gzip":    NewGzip,
			"deflate": NewDeflate,
		},
		Types:    []string{"text/*"},
		ErrorLog: log.New(os.Stderr, "", log.LstdFlags),
	}
	srv := rest.NewServer(t, New(http.HandlerFunc(f4), opt), nil)

	// 指定 accept-encoding = *
	srv.NewRequest(http.MethodGet, "/").
		Header("Accept-Encoding", "*").
		Do().
		Status(http.StatusAccepted).
		StringBody("f1\nf2").
		Header("Content-Encoding", "")

	// 指定 accept-encoding = identity
	srv.NewRequest(http.MethodGet, "/").
		Header("Accept-Encoding", "identity").
		Do().
		Status(http.StatusAccepted).
		StringBody("f1\nf2").
		Header("Content-Encoding", "")

	// 指定 accept-encoding 为空
	srv.NewRequest(http.MethodGet, "/").
		Header("Accept-Encoding", "").
		Do().
		Status(http.StatusAccepted).
		StringBody("f1\nf2").
		Header("Content-Encoding", "")

	// accept-encoding = deflate
	buf := new(bytes.Buffer)
	srv.NewRequest(http.MethodGet, "/").
		Header("Accept-encoding", "gzip;q=0.8,deflate").
		Do().
		Status(http.StatusAccepted).
		BodyNotNil().
		ReadBody(buf).
		Header("Content-Type", "text/plain; charset=utf-8").
		Header("Content-Encoding", "").
		Header("Vary", "")
}

func TestCompress_empty(t *testing.T) {
	a := assert.New(t)

	opt := &Options{
		Funcs: map[string]WriterFunc{
			"gzip":    NewGzip,
			"deflate": NewDeflate,
		},
		Types:    []string{"text/*"},
		ErrorLog: log.New(os.Stderr, "", log.LstdFlags),
	}

	// 不输出任何信息
	f5 := func(w http.ResponseWriter, r *http.Request) {}
	srv := rest.NewServer(t, New(http.HandlerFunc(f5), opt), nil)
	// accept-encoding = deflate
	buf := new(bytes.Buffer)
	srv.NewRequest(http.MethodGet, "/").
		Header("Accept-encoding", "gzip;q=0.8,deflate").
		Do().
		Status(http.StatusOK).
		ReadBody(buf).
		Header("Content-Type", "").
		Header("Content-Encoding", "").
		Header("Vary", "")
	a.Equal(0, buf.Len())

	// 不输出任何信息
	f5 = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}
	srv = rest.NewServer(t, New(http.HandlerFunc(f5), opt), nil)
	// accept-encoding = deflate
	buf = new(bytes.Buffer)
	srv.NewRequest(http.MethodGet, "/").
		Header("Accept-encoding", "gzip;q=0.8,deflate").
		Do().
		ReadBody(buf).
		Status(http.StatusAccepted).
		Header("Content-Type", "").
		Header("Content-Encoding", "").
		Header("Vary", "")
	a.Equal(0, buf.Len())
}
