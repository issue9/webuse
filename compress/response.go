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
	writer         io.Writer
	compressWriter io.WriteCloser
	responseWriter http.ResponseWriter
	wroteHeader    bool

	c            *Compress
	f            WriterFunc
	encodingName string
}

func (c *Compress) newResponse(resp http.ResponseWriter, f WriterFunc, encodingName string) *response {
	return &response{
		responseWriter: resp,
		c:              c,
		f:              f,
		encodingName:   encodingName,
	}
}

func (resp *response) Header() http.Header {
	return resp.responseWriter.Header()
}

// 根据接口要求：一旦调用此函数，之后产生的报头将不再启作用。
func (resp *response) WriteHeader(code int) {
	resp.write(code, nil)
}

// NOTE: 根据接口要求，第一次调用 Write 时，会发送报头内容，
// 即 WriteHeader(200) 自动调用，即使写入的是空内容。
func (resp *response) Write(bs []byte) (int, error) {
	if !resp.wroteHeader {
		resp.write(http.StatusOK, bs)
	}
	return resp.writer.Write(bs)
}

func (resp *response) write(status int, bs []byte) {
	defer func() {
		resp.responseWriter.WriteHeader(status)
		resp.wroteHeader = true
	}()

	h := resp.Header()

	// https://github.com/golang/go/issues/14975
	h.Del("Content-Length")

	ct := h.Get("Content-Type")
	if ct == "" {
		ct = http.DetectContentType(bs)
		h.Set("Content-Type", ct)
	}
	if !bodyAllowedForStatus(status) || !resp.c.canCompressed(ct) {
		resp.writer = resp.responseWriter
		return
	}

	if compressWriter, err := resp.f(resp.responseWriter); err != nil { // 转换出错，退化成 responseWriter
		resp.c.printError(err)
		resp.writer = resp.responseWriter
	} else {
		h.Set("Content-Encoding", resp.encodingName)
		h.Add("Vary", "Content-Encoding")
		resp.compressWriter = compressWriter
		resp.writer = resp.compressWriter
	}
}

func (resp *response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := resp.responseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	panic("未实现 http.Hijacker")
}

func (resp *response) close() {
	if resp.compressWriter != nil {
		if err := resp.compressWriter.Close(); err != nil {
			resp.c.printError(err)
		}
	}
}

// 以下内容复制于官方标准库
//
// bodyAllowedForStatus reports whether a given response status code
// permits a body. See RFC 7230, section 3.3.
func bodyAllowedForStatus(status int) bool {
	switch {
	case status >= 100 && status <= 199:
		return false
	case status == 204:
		return false
	case status == 304:
		return false
	}
	return true
}
