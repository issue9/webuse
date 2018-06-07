// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package compress

import (
	"bufio"
	"io"
	"net"
	"net/http"
)

// 实现了 http.ResponseWriter 接口。
type response struct {
	gzw io.WriteCloser      // 实现压缩功能的 io.Writer
	rw  http.ResponseWriter // 旧的 ResponseWriter
}

func (resp *response) Write(bs []byte) (int, error) {
	h := resp.rw.Header()
	if h.Get("Content-Type") == "" {
		h.Set("Content-Type", http.DetectContentType(bs))
	}

	return resp.gzw.Write(bs)
}

func (resp *response) Header() http.Header {
	return resp.rw.Header()
}

func (resp *response) WriteHeader(code int) {
	// https://github.com/golang/go/issues/14975
	resp.rw.Header().Del("Content-Length")

	resp.rw.WriteHeader(code)
}

func (resp *response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := resp.rw.(http.Hijacker); ok {
		return hj.Hijack()
	}

	panic("未实现 http.Hijacker")
}
