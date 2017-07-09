// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package compress 提供一个支持内容压缩的中间件。
package compress

import (
	"bufio"
	"compress/flate"
	"compress/gzip"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

// 实现了 http.ResponseWriter 接口。
type compressWriter struct {
	gzw io.Writer
	rw  http.ResponseWriter
	hj  http.Hijacker
}

func (cw *compressWriter) Write(bs []byte) (int, error) {
	h := cw.rw.Header()
	if h.Get("Content-Type") == "" {
		h.Set("Content-Type", http.DetectContentType(bs))
	}

	return cw.gzw.Write(bs)
}

func (cw *compressWriter) Header() http.Header {
	return cw.rw.Header()
}

func (cw *compressWriter) WriteHeader(code int) {
	// https://github.com/golang/go/issues/14975
	cw.rw.Header().Del("Content-Length")

	cw.rw.WriteHeader(code)
}

func (cw *compressWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if cw.hj != nil {
		return cw.hj.Hijack()
	}

	panic("未实现 http.Hijacker")
}

type compress struct {
	h      http.Handler
	errlog *log.Logger
}

// New 构建一个支持压缩的中间件。
// 支持 gzip 或是 deflate 功能的 handler。
// 根据客户端请求内容自动匹配相应的压缩算法，优先匹配 gzip。
//
// NOTE: 经过压缩的内容，可能需要重新指定 Content-Type，系统检测的类型未必正确。
func New(next http.Handler, errlog *log.Logger) http.Handler {
	return &compress{
		h:      next,
		errlog: errlog,
	}
}

func (c *compress) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var gzw io.WriteCloser
	var encoding string
	encodings := strings.Split(r.Header.Get("Accept-Encoding"), ",")
	for _, encoding = range encodings {
		encoding = strings.ToLower(strings.TrimSpace(encoding))

		if encoding == "gzip" {
			gzw = gzip.NewWriter(w)
			break
		}

		if encoding == "deflate" {
			var err error
			gzw, err = flate.NewWriter(w, flate.DefaultCompression)
			if err != nil { // 若出错，不压缩，直接返回
				c.errlog.Println(err)
				c.h.ServeHTTP(w, r)
				return
			}
			break
		}
	} // end for
	if gzw == nil { // 不支持的压缩格式
		return
	}

	w.Header().Set("Content-Encoding", encoding)
	w.Header().Add("Vary", "Accept-Encoding")
	cw := &compressWriter{
		gzw: gzw,
		rw:  w,
	}
	if hj, ok := w.(http.Hijacker); ok { // http2 没有实现该接口
		cw.hj = hj
	}

	defer gzw.Close() // 只要 gzw!=nil 的，必须会执行到此处。

	// 此处可能 panic，所以得保证在 panic 之前，gzw 变量已经 Close
	c.h.ServeHTTP(cw, r)
}
