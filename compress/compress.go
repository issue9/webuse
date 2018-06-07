// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package compress 提供一个支持内容压缩的中间件。
package compress

import (
	"compress/flate"
	"compress/gzip"
	"io"
	"log"
	"net/http"

	"github.com/issue9/middleware/compress/accept"
)

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
	accepts, err := accept.Parse(r.Header.Get("Accept-Encoding"))
	if err != nil {
		c.errlog.Println(err)
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	var gzw io.WriteCloser
	var accept *accept.Accept
	for _, accept = range accepts {
		if accept.Value == "gzip" {
			gzw = gzip.NewWriter(w)
			break
		}

		if accept.Value == "deflate" {
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

	w.Header().Set("Content-Encoding", accept.Value)
	w.Header().Add("Vary", "Accept-Encoding")
	resp := &response{
		gzw: gzw,
		rw:  w,
	}

	defer gzw.Close() // 只要 gzw!=nil 的，必须会执行到此处。

	// 此处可能 panic，所以得保证在 panic 之前，gzw 变量已经 Close
	c.h.ServeHTTP(resp, r)
}
