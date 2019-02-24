// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package compress

import (
	"bufio"
	"bytes"
	"net"
	"net/http"
)

// 实现了 http.ResponseWriter 接口。
type response struct {
	// 所有返回给客户端的内容，先保存到 buffer。
	// 在最终返回给客户之前，即 close() 函数中，
	// 再判断其是否符合压缩的条件，根据条件输出到 rw 或是压缩对象。
	buffer *bytes.Buffer

	rw           http.ResponseWriter // 旧的 ResponseWriter
	opt          *Options
	f            WriterFunc
	encodingName string
}

func (resp *response) Write(bs []byte) (int, error) {
	return resp.buffer.Write(bs)
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

func (resp *response) close() {
	bs := resp.buffer.Bytes()
	h := resp.Header()

	hv := h.Get("Content-Type")
	if hv == "" {
		h.Set("Content-Type", http.DetectContentType(bs))
	}

	if !resp.opt.canCompressed(resp.buffer.Len(), h.Get("Content-Type")) {
		if _, err := resp.rw.Write(bs); err != nil {
			resp.opt.println(err)
		}
		return
	}

	gzw, err := resp.f(resp.rw)
	if err != nil {
		resp.opt.println(err)

		if _, err := resp.rw.Write(bs); err != nil {
			resp.opt.println(err)
		}
		return
	}

	h.Set("Content-Encoding", resp.encodingName)
	h.Add("Vary", "Content-Encoding")

	if _, err = gzw.Write(bs); err != nil {
		resp.opt.println(err)
		return
	}

	if err = gzw.Close(); err != nil {
		resp.opt.println(err)
	}
}
