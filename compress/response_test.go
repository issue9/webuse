// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package compress

import (
	"bytes"
	"compress/flate"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

var (
	_ http.ResponseWriter = &response{}
	_ http.Hijacker       = &response{}

	_ WriterFunc = newErrorWriter
)

func newErrorWriter(w io.Writer) (io.WriteCloser, error) {
	return nil, errors.New("error")
}

func TestResponse_close(t *testing.T) {
	a := assert.New(t)
	rw := httptest.NewRecorder()
	opt := &Options{
		Funcs: map[string]WriterFunc{"deflate": NewDeflate},
		Types: []string{"application/xml", "text/*", "application/json"},
		Size:  10,
	}
	opt.build()
	resp := &response{
		rw:           rw,
		buffer:       new(bytes.Buffer),
		opt:          opt,
		f:            NewDeflate,
		encodingName: "deflate",
	}

	_, err := resp.Write([]byte("123"))
	a.NotError(err)
	resp.close()
	a.Equal(resp.buffer.String(), "123").
		Equal(resp.buffer.String(), rw.Body.String()).
		Equal(rw.Header().Get("Content-Encoding"), "")

	// 多次写入
	resp.buffer.Reset()
	rw = httptest.NewRecorder()
	resp.rw = rw
	_, err = resp.Write([]byte("123"))
	a.NotError(err)
	_, err = resp.Write([]byte("4567890\n"))
	a.NotError(err)
	_, err = resp.Write([]byte("123"))
	a.NotError(err)
	resp.close()
	a.Equal(resp.buffer.String(), "1234567890\n123").
		NotEqual(resp.buffer.String(), rw.Body.String()).
		Equal(rw.Header().Get("Content-Encoding"), "deflate")
	data, err := ioutil.ReadAll(flate.NewReader(rw.Body))
	a.NotError(err).NotNil(data).
		Equal(string(data), resp.buffer.String())

	// 可压缩，但是压缩时构建压缩实例出错
	resp.buffer.Reset()
	rw = httptest.NewRecorder()
	resp.rw = rw
	resp.f = newErrorWriter
	_, err = resp.Write([]byte("1234567890\n123"))
	a.NotError(err)
	resp.close()
	a.Equal(resp.buffer.String(), "1234567890\n123").
		Equal(resp.buffer.String(), rw.Body.String()).
		Equal(rw.Header().Get("Content-Encoding"), "")
}
