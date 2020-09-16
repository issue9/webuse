// SPDX-License-Identifier: MIT

package compress

import (
	"bufio"
	"io"
	"net"
	"net/http"
)

// 实现了 http.ResponseWriter 接口
type response struct {
	// 当前的 Write 方法实际调用的对象
	//
	// 可能是 rw 也有可能是 gzw，根据第一次调用 Write
	// 时，判断引用哪个对象。
	writer io.Writer

	// gzw 是根据当前的 f 值生成的压缩对象实例。
	gzw io.WriteCloser

	rw           http.ResponseWriter // 旧的 ResponseWriter
	c            *Compress
	f            WriterFunc
	encodingName string
}

func (resp *response) Header() http.Header {
	return resp.rw.Header()
}

// 根据接口要求：一旦调用此函数，之后产生的报头将不再启作用。
func (resp *response) WriteHeader(code int) {
	// https://github.com/golang/go/issues/14975
	resp.rw.Header().Del("Content-Length")

	resp.genWriter(resp.Header().Get("Content-Type"))
	resp.rw.WriteHeader(code)
}

// 根据接口要求，第一次调用 Write 时，会发送报头内容，即 WriteHeader 自动调用。
func (resp *response) Write(bs []byte) (int, error) {
	if resp.writer == nil {
		h := resp.Header()

		ct := h.Get("Content-Type")
		if ct == "" {
			ct = http.DetectContentType(bs)
			h.Set("Content-Type", ct)
		}

		resp.genWriter(ct)
	}

	return resp.writer.Write(bs)
}

func (resp *response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := resp.rw.(http.Hijacker); ok {
		return hj.Hijack()
	}

	panic("未实现 http.Hijacker")
}

func (resp *response) close() {
	if resp.gzw != nil {
		if err := resp.gzw.Close(); err != nil {
			resp.c.printError(err)
		}
	}
}

func (resp *response) genWriter(ct string) {
	h := resp.Header()

	if !resp.c.canCompressed(ct) {
		resp.writer = resp.rw
		return
	}

	if gzw, err := resp.f(resp.rw); err != nil {
		resp.c.printError(err)
		resp.writer = resp.rw
	} else {
		h.Set("Content-Encoding", resp.encodingName)
		h.Add("Vary", "Content-Encoding")
		resp.gzw = gzw
		resp.writer = resp.gzw
	}
}
