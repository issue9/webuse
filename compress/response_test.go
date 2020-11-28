// SPDX-License-Identifier: MIT

package compress

import (
	"compress/flate"
	"compress/gzip"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/issue9/assert"
)

var (
	_ http.ResponseWriter = &response{}
	_ http.Hijacker       = &response{}

	_ WriterFunc = newErrorWriter
)

func newErrorWriter(w io.Writer) (Writer, error) {
	return nil, errors.New("error")
}

func TestResponse_Write(t *testing.T) {
	a := assert.New(t)
	rw := httptest.NewRecorder()

	c := New(log.New(os.Stderr, "", log.LstdFlags), map[string]WriterFunc{
		"deflate": NewDeflate,
	}, "application/xml", "text/*", "application/json")
	a.NotNil(c)

	resp := c.newResponse(rw, NewDeflate, "deflate")

	// 压缩
	_, err := resp.Write([]byte("123"))
	a.NotError(err)
	resp.close()
	a.NotEmpty(rw.Body.String()).
		Equal(rw.Header().Get("Content-Encoding"), "deflate").
		Equal(rw.Code, http.StatusOK)

	data, err := ioutil.ReadAll(flate.NewReader(rw.Body))
	a.NotError(err).NotNil(data).
		Equal(string(data), "123")

	// 没有写入内容
	rw = httptest.NewRecorder()
	resp = c.newResponse(rw, NewDeflate, "deflate")
	resp.close()
	a.Empty(rw.Body.String()).
		Equal(rw.Header().Get("Content-Encoding"), "").
		Equal(rw.Code, http.StatusOK)

	// 写入空内容
	rw = httptest.NewRecorder()
	resp = c.newResponse(rw, NewDeflate, "deflate")
	n, err := resp.Write(nil)
	a.NotError(err).Equal(0, n)
	resp.close()
	a.NotEmpty(rw.Body.String()).
		Equal(rw.Header().Get("Content-Encoding"), "deflate").
		Equal(rw.Code, http.StatusOK)

	// 多次写入
	rw = httptest.NewRecorder()
	resp = c.newResponse(rw, NewDeflate, "deflate")
	_, err = resp.Write([]byte("123"))
	a.NotError(err)
	_, err = resp.Write([]byte("4567890\n"))
	a.NotError(err)
	_, err = resp.Write([]byte("123"))
	a.NotError(err)
	resp.close()
	a.NotEmpty(rw.Body.String())

	data, err = ioutil.ReadAll(flate.NewReader(rw.Body))
	a.NotError(err).NotNil(data).
		Equal(string(data), "1234567890\n123")

	// 可压缩，但是压缩时构建压缩实例出错
	rw = httptest.NewRecorder()
	resp = c.newResponse(rw, newErrorWriter, "deflate")
	resp.f = newErrorWriter
	_, err = resp.Write([]byte("1234567890\n123"))
	a.NotError(err)
	resp.close()
	a.Equal(rw.Body.String(), "1234567890\n123").
		Equal(rw.Header().Get("Content-Encoding"), "")

	// 可压缩
	rw = httptest.NewRecorder()
	resp = c.newResponse(rw, NewGzip, "deflate")
	_, err = resp.Write([]byte("1234567890\n123"))
	a.NotError(err)
	resp.close()
	a.NotEqual(rw.Body.String(), "1234567890\n123").
		Equal(rw.Header().Get("Content-Encoding"), resp.encodingName)

	gzw, err := gzip.NewReader(rw.Body)
	a.NotError(err).NotNil(gzw)
	data, err = ioutil.ReadAll(gzw)
	a.NotError(err).NotNil(data).
		Equal(string(data), "1234567890\n123")
}
