// SPDX-License-Identifier: MIT

package compress

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"sync"
)

var respPool = &sync.Pool{
	New: func() interface{} { return &response{} },
}

// 实现了 http.ResponseWriter 接口
type response struct {
	c *Compress

	writer         io.Writer
	responseWriter http.ResponseWriter
	wroteHeader    bool

	// 压缩相关的字段
	//
	// 如果 f 为 nil，表示不需要压缩
	f            Writer
	encodingName string
}

func (c *Compress) newResponse(resp http.ResponseWriter, f Writer, encodingName string) *response {
	r := respPool.Get().(*response)
	r.c = c
	r.writer = nil
	r.responseWriter = resp
	r.wroteHeader = false
	r.f = f
	r.encodingName = encodingName
	return r
}

func (resp *response) Header() http.Header { return resp.responseWriter.Header() }

func (resp *response) WriteHeader(code int) {
	if !resp.wroteHeader {
		resp.writeHeader(code, nil)
	}
}

// NOTE: 根据接口要求，第一次调用 Write 时，会发送报头内容，
// 即 WriteHeader(200) 自动调用，即使写入的是空内容。
func (resp *response) Write(bs []byte) (int, error) {
	if !resp.wroteHeader {
		resp.writeHeader(http.StatusOK, bs)
	}
	return resp.writer.Write(bs)
}

// bs 仅作为检测 content-type 类型值
func (resp *response) writeHeader(status int, bs []byte) {
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
	if resp.f == nil || !bodyAllowedForStatus(status) || !resp.c.canCompressed(ct) {
		resp.writer = resp.responseWriter
		resp.c.putAlgorithm(resp.encodingName, resp.f)
		resp.f = nil
		return
	}

	resp.f.Reset(resp.responseWriter)
	h.Set("Content-Encoding", resp.encodingName)
	h.Add("Vary", "Content-Encoding")
	resp.writer = resp.f
}

func (resp *response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := resp.responseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	panic("未实现 http.Hijacker")
}

func (resp *response) close() {
	if resp.f != nil {
		if err := resp.f.Close(); err != nil {
			resp.c.printError(err)
		}
		resp.c.putAlgorithm(resp.encodingName, resp.f)
		resp.f = nil
	}

	respPool.Put(resp)
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
