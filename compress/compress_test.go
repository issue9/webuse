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

	"github.com/andybalholm/brotli"
	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"
)

var (
	_ WriterFunc = NewDeflate
	_ WriterFunc = NewGzip
	_ WriterFunc = NewBrotli
)

var (
	f1 = func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("f1\nf2"))
	}

	f2 = func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("f1\nf2"))
	}

	f3 = func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("f1\nf2"))
	}

	f4 = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("f1\nf2"))
	}

	f5 = func(w http.ResponseWriter, r *http.Request) {
		// 在 204 状态下写入空值，但是压缩格式本身是带内容的，一旦输出该内容，
		// 会报 http.ErrBodyNotAllowed 的错误。
		w.WriteHeader(http.StatusNoContent)
		w.Write(nil)
	}

	// 设置了内容类型，但是实际未输出任何内容
	f6 = func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write(nil)
	}

	// 输出空内容
	f7 = func(w http.ResponseWriter, r *http.Request) {
		w.Write(nil)
	}
)

func TestNew(t *testing.T) {
	a := assert.New(t)

	c := New(log.New(os.Stderr, "", log.LstdFlags), map[string]WriterFunc{
		"gzip": NewGzip,
	}, "application/xml", "text/*", "application/json")
	a.NotNil(c)

	a.Equal(c.prefix, []string{"text/"})
	a.Equal(c.types, []string{"application/xml", "application/json"})
}

func TestCompress_SetWriter(t *testing.T) {
	a := assert.New(t)

	c := New(log.New(os.Stderr, "", log.LstdFlags), map[string]WriterFunc{
		"gzip": NewGzip,
	}, "application/xml", "text/*", "application/json")
	a.NotNil(c)

	a.Equal(1, len(c.writers))
	c.SetWriter("gzip", nil)
	a.Equal(0, len(c.writers))

	c.SetWriter("gzip", NewGzip)
	c.SetWriter("br", NewBrotli)
	a.Equal(2, len(c.writers))
}

func TestCompress_Types(t *testing.T) {
	a := assert.New(t)

	c := New(log.New(os.Stderr, "", log.LstdFlags), map[string]WriterFunc{
		"gzip": NewGzip,
	}, "application/xml", "text/*", "application/json")
	a.NotNil(c)

	a.Equal(2, len(c.types)).
		Equal(1, len(c.prefix)).
		False(c.any)

	c.DeleteType("application/")
	a.Equal(2, len(c.types)).
		Equal(1, len(c.prefix)).
		False(c.any)

	c.DeleteType("application/*")
	a.Equal(0, len(c.types)).
		Equal(1, len(c.prefix)).
		False(c.any)

	c.DeleteType("*")
	a.Equal(0, len(c.types)).
		Equal(0, len(c.prefix)).
		False(c.any)
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
	a.NotNil(c)

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

	// *
	c = New(log.New(os.Stderr, "", log.LstdFlags), map[string]WriterFunc{
		"gzip":    NewGzip,
		"deflate": NewDeflate,
	}, "*")
	a.NotNil(c)

	srv = rest.NewServer(t, c.MiddlewareFunc(f1), nil)
	defer srv.Close()

	// accept-encoding = gzip
	srv.NewRequest(http.MethodGet, "/").
		Header("Accept-encoding", "gzip,deflate;q=0.8").
		Do().
		BodyNotNil().
		Status(http.StatusAccepted).
		Header("Content-Type", "text/html").
		Header("Content-Encoding", "gzip").
		Header("Vary", "Content-Encoding")
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

	// *
	c = New(log.New(os.Stderr, "", log.LstdFlags), map[string]WriterFunc{
		"gzip":    NewGzip,
		"deflate": NewDeflate,
	}, "*")
	a.NotNil(c)

	srv = rest.NewServer(t, c.MiddlewareFunc(f2), nil)
	defer srv.Close()

	// accept-encoding = deflate
	srv.NewRequest(http.MethodGet, "/").
		Header("Accept-encoding", "gzip,deflate;q=0.8").
		Do().
		Status(http.StatusOK).
		BodyNotNil().
		Header("Content-Type", "text/html").
		Header("Content-Encoding", "gzip").
		Header("Vary", "Content-Encoding")
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

	// accept-encoding = br
	buf := new(bytes.Buffer)
	srv.NewRequest(http.MethodGet, "/").
		Header("Accept-encoding", "gzip;q=0.8,br").
		Do().
		Status(http.StatusOK).
		BodyNotNil().
		ReadBody(buf).
		Header("Content-Type", "text/plain; charset=utf-8").
		Header("Content-Encoding", "br").
		Header("Vary", "Content-Encoding")

	// 解码后相等
	a.True(len(buf.Bytes()) > 0)
	data, err := ioutil.ReadAll(brotli.NewReader(buf))
	a.NotError(err).NotNil(data)
	a.Equal(string(data), "f1\nf2")

	// *
	c = New(log.New(os.Stderr, "", log.LstdFlags), map[string]WriterFunc{
		"gzip":    NewGzip,
		"deflate": NewDeflate,
	}, "*")
	a.NotNil(c)

	srv = rest.NewServer(t, c.MiddlewareFunc(f3), nil)
	defer srv.Close()

	// accept-encoding = deflate
	srv.NewRequest(http.MethodGet, "/").
		Header("Accept-encoding", "gzip;q=0.8,deflate").
		Do().
		Status(http.StatusOK).
		BodyNotNil().
		Header("Content-Type", "text/plain; charset=utf-8").
		Header("Content-Encoding", "deflate").
		Header("Vary", "Content-Encoding")
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
		Header("Content-Type", "text/plain; charset=utf-8"). // 未动用压缩，在调用默认的 Write 时会检测内容类型
		Header("Content-Encoding", "").
		Header("Vary", "")

	// *
	c = New(log.New(os.Stderr, "", log.LstdFlags), map[string]WriterFunc{
		"gzip":    NewGzip,
		"deflate": NewDeflate,
	}, "*")
	a.NotNil(c)

	srv = rest.NewServer(t, c.MiddlewareFunc(f4), nil)
	defer srv.Close()

	srv.NewRequest(http.MethodGet, "/").
		Header("Accept-encoding", "gzip;q=0.8,deflate").
		Do().
		Status(http.StatusAccepted).
		BodyNotNil().
		ReadBody(buf).
		// 启用了压缩，此时还不知道类型，所以此值可能为空，
		// go 1.13 会被此值默认设置为 application/octet-stream。
		// f4 的实现是先调用 WriteHeader，一旦调用之后的报头输出也不再启作用。
		//Header("Content-Type", "").
		Header("Content-Encoding", "deflate").
		Header("Vary", "Content-Encoding")
}

func TestCompress_f5(t *testing.T) {
	a := assert.New(t)

	c := New(log.New(os.Stderr, "", log.LstdFlags), map[string]WriterFunc{
		"gzip":    NewGzip,
		"deflate": NewDeflate,
		"br":      NewBrotli,
	}, "text/*")
	a.NotNil(c)

	srv := rest.NewServer(t, c.MiddlewareFunc(f5), nil)

	// 指定 accept-encoding = *
	srv.NewRequest(http.MethodGet, "/").
		Header("Accept-Encoding", "*").
		Do().
		Status(http.StatusNoContent).
		Header("Content-Encoding", "")

	// 指定 accept-encoding = identity
	srv.NewRequest(http.MethodGet, "/").
		Header("Accept-Encoding", "identity").
		Do().
		Status(http.StatusNoContent).
		Header("Content-Encoding", "")

	// 指定 accept-encoding 为空
	srv.NewRequest(http.MethodGet, "/").
		Header("Accept-Encoding", "").
		Do().
		Status(http.StatusNoContent).
		Header("Content-Encoding", "")

	// accept-encoding = deflate
	srv.NewRequest(http.MethodGet, "/").
		Header("Accept-encoding", "gzip;q=0.8,deflate").
		Do().
		Status(http.StatusNoContent).
		Header("Content-Encoding", "").
		Header("Vary", "")

	// *
	c = New(log.New(os.Stderr, "", log.LstdFlags), map[string]WriterFunc{
		"gzip":    NewGzip,
		"deflate": NewDeflate,
	}, "*")
	a.NotNil(c)

	srv = rest.NewServer(t, c.MiddlewareFunc(f5), nil)
	defer srv.Close()

	srv.NewRequest(http.MethodGet, "/").
		Header("Accept-encoding", "gzip;q=0.8,deflate").
		Do().
		Status(http.StatusNoContent).
		BodyEmpty().
		Header("Content-Encoding", "").
		Header("Vary", "")
}

func TestCompress_f6(t *testing.T) {
	a := assert.New(t)

	c := New(log.New(os.Stderr, "", log.LstdFlags), map[string]WriterFunc{
		"gzip":    NewGzip,
		"deflate": NewDeflate,
		"br":      NewBrotli,
	}, "text/*")
	a.NotNil(c)

	srv := rest.NewServer(t, c.MiddlewareFunc(f6), nil)

	// 指定 accept-encoding = *
	srv.NewRequest(http.MethodGet, "/").
		Header("Accept-Encoding", "*").
		Do().
		Status(http.StatusOK).
		Header("Content-Encoding", "")

	// 指定 accept-encoding = identity
	srv.NewRequest(http.MethodGet, "/").
		Header("Accept-Encoding", "identity").
		Do().
		Status(http.StatusOK).
		Header("Content-Encoding", "")

	// 指定 accept-encoding 为空
	srv.NewRequest(http.MethodGet, "/").
		Header("Accept-Encoding", "").
		Do().
		Status(http.StatusOK).
		Header("Content-Encoding", "")

	// accept-encoding = deflate
	srv.NewRequest(http.MethodGet, "/").
		Header("Accept-encoding", "gzip;q=0.8,deflate").
		Do().
		Status(http.StatusOK).
		Header("Content-Encoding", "").
		Header("Vary", "")

	// *
	c = New(log.New(os.Stderr, "", log.LstdFlags), map[string]WriterFunc{
		"gzip":    NewGzip,
		"deflate": NewDeflate,
	}, "*")
	a.NotNil(c)

	srv = rest.NewServer(t, c.MiddlewareFunc(f6), nil)
	defer srv.Close()

	srv.NewRequest(http.MethodGet, "/").
		Header("Accept-encoding", "gzip;q=0.8,deflate").
		Do().
		Status(http.StatusOK).
		BodyEmpty().
		Header("Content-Encoding", "").
		Header("Vary", "")
}

func TestCompress_f7(t *testing.T) {
	a := assert.New(t)

	c := New(log.New(os.Stderr, "", log.LstdFlags), map[string]WriterFunc{
		"gzip":    NewGzip,
		"deflate": NewDeflate,
		"br":      NewBrotli,
	}, "text/*")
	a.NotNil(c)

	srv := rest.NewServer(t, c.MiddlewareFunc(f7), nil)

	// 指定 accept-encoding = *
	srv.NewRequest(http.MethodGet, "/").
		Header("Accept-Encoding", "*").
		Do().
		Status(http.StatusOK).
		Header("Content-Encoding", "")

	// 指定 accept-encoding = identity
	srv.NewRequest(http.MethodGet, "/").
		Header("Accept-Encoding", "identity").
		Do().
		Status(http.StatusOK).
		Header("Content-Encoding", "")

	// 指定 accept-encoding 为空
	srv.NewRequest(http.MethodGet, "/").
		Header("Accept-Encoding", "").
		Do().
		Status(http.StatusOK).
		Header("Content-Encoding", "")

	// accept-encoding = deflate
	srv.NewRequest(http.MethodGet, "/").
		Header("Accept-encoding", "gzip;q=0.8,deflate").
		Do().
		Status(http.StatusOK).
		Header("Content-Encoding", "").
		Header("Vary", "")

	// *
	c = New(log.New(os.Stderr, "", log.LstdFlags), map[string]WriterFunc{
		"gzip":    NewGzip,
		"deflate": NewDeflate,
	}, "*")
	a.NotNil(c)

	srv = rest.NewServer(t, c.MiddlewareFunc(f7), nil)
	defer srv.Close()

	srv.NewRequest(http.MethodGet, "/").
		Header("Accept-encoding", "gzip;q=0.8,deflate").
		Do().
		Status(http.StatusOK).
		BodyEmpty().
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
