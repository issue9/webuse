// SPDX-License-Identifier: MIT

package compress

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

var (
	_ http.ResponseWriter = &response{}
	_ http.Hijacker       = &response{}
)

func newWriter(a *assert.Assertion, f WriterFunc) Writer {
	w, err := f(new(bytes.Buffer))
	a.NotError(err).NotNil(w)
	return w
}

func TestResponse_Write(t *testing.T) {
	a := assert.New(t)
	rw := httptest.NewRecorder()

	c := newCompress(a, "application/xml", "text/*", "application/json")

	resp := c.newResponse(rw, newWriter(a, NewDeflate), "deflate")

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
	resp = c.newResponse(rw, newWriter(a, NewDeflate), "deflate")
	resp.close()
	a.Empty(rw.Body.String()).
		Equal(rw.Header().Get("Content-Encoding"), "").
		Equal(rw.Code, http.StatusOK)

	// 写入空内容
	rw = httptest.NewRecorder()
	resp = c.newResponse(rw, newWriter(a, NewDeflate), "deflate")
	n, err := resp.Write(nil)
	a.NotError(err).Equal(0, n)
	resp.close()
	a.NotEmpty(rw.Body.String()).
		Equal(rw.Header().Get("Content-Encoding"), "deflate").
		Equal(rw.Code, http.StatusOK)

	// 多次写入
	rw = httptest.NewRecorder()
	resp = c.newResponse(rw, newWriter(a, NewDeflate), "deflate")
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

	// 可压缩
	rw = httptest.NewRecorder()
	resp = c.newResponse(rw, newWriter(a, NewGzip), "deflate")
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

func TestBodyAllowedForStatus(t *testing.T) {
	a := assert.New(t)

	a.True(bodyAllowedForStatus(http.StatusAccepted))
	a.True(bodyAllowedForStatus(http.StatusOK))
	a.False(bodyAllowedForStatus(http.StatusNoContent))
	a.False(bodyAllowedForStatus(100))
}
