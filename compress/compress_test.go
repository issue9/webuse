// SPDX-License-Identifier: MIT

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

func TestNew(t *testing.T) {
	a := assert.New(t)

	c := New(log.New(os.Stderr, "", log.LstdFlags), map[string]WriterFunc{
		"gzip": NewGzip,
	}, "application/xml", "text/*", "application/json")
	a.NotNil(c)

	a.Equal(c.prefix, []string{"text/"})
	a.Equal(c.types, []string{"application/xml", "application/json"})
}

func TestCompress_f1(t *testing.T) {
	a := assert.New(t)

	c := New(nil, nil)
	a.NotNil(c)

	// 空的 options
	buf := new(bytes.Buffer)
	srv := rest.NewServer(t, c.MiddlewareFunc(f1), nil)
	srv.NewRequest(http.MethodGet, "/").
		Header("Accept-encoding", "gzip,deflate;q=0.8").
		Do().
		BodyNotNil().
		ReadBody(buf).
		Status(http.StatusAccepted).
		Header("Content-Type", "text/html").
		Header("Content-Encoding", "").
		Header("Vary", "")
	a.Equal(buf.String(), "f1\nf2")
	srv.Close()

	c = New(log.New(os.Stderr, "", log.LstdFlags), map[string]WriterFunc{
		"gzip":    NewGzip,
		"deflate": NewDeflate,
	}, "text/*")

	srv = rest.NewServer(t, c.MiddlewareFunc(f1), nil)
	defer srv.Close()

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
	buf = new(bytes.Buffer)
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

	c := New(log.New(os.Stderr, "", log.LstdFlags), map[string]WriterFunc{
		"gzip":    NewGzip,
		"deflate": NewDeflate,
	}, "text/*")
	a.NotNil(c)

	srv := rest.NewServer(t, c.MiddlewareFunc(f2), nil)

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

	c := New(log.New(os.Stderr, "", log.LstdFlags), map[string]WriterFunc{
		"gzip":    NewGzip,
		"deflate": NewDeflate,
		"br":      NewBrotli,
	}, "text/*")
	a.NotNil(c)

	srv := rest.NewServer(t, c.MiddlewareFunc(f3), nil)

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
	a := assert.New(t)

	c := New(log.New(os.Stderr, "", log.LstdFlags), map[string]WriterFunc{
		"gzip":    NewGzip,
		"deflate": NewDeflate,
		"br":      NewBrotli,
	}, "text/*")
	a.NotNil(c)

	srv := rest.NewServer(t, c.MiddlewareFunc(f4), nil)

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

	c := New(log.New(os.Stderr, "", log.LstdFlags), map[string]WriterFunc{
		"gzip":    NewGzip,
		"deflate": NewDeflate,
		"br":      NewBrotli,
	}, "text/*")
	a.NotNil(c)

	// 不输出任何信息
	f5 := func(w http.ResponseWriter, r *http.Request) {}
	srv := rest.NewServer(t, c.MiddlewareFunc(f5), nil)
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
	srv = rest.NewServer(t, c.MiddlewareFunc(f5), nil)
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

func TestCompress_canCompressed(t *testing.T) {
	a := assert.New(t)

	c := New(nil, nil)
	a.NotNil(c)

	a.False(c.canCompressed(""))

	c = New(log.New(os.Stderr, "", log.LstdFlags), map[string]WriterFunc{
		"gzip": NewGzip,
	}, "text/*", "application/json")
	a.NotNil(c)

	// 长度不够
	a.False(c.canCompressed(""))

	// 长度够，但是未指定 content-type
	a.False(c.canCompressed(""))

	a.True(c.canCompressed("text/html;charset=utf-8"))

	a.True(c.canCompressed("application/json"))

	a.False(c.canCompressed("application/octet"))
}
