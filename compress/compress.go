// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package compress 提供一个支持内容压缩的中间件。
package compress

import (
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"

	"github.com/issue9/middleware/compress/accept"
)

// WriterFunc 定义了将一个 io.Writer 声明为具有压缩功能的 io.WriteCloser
type WriterFunc func(w io.Writer) (io.WriteCloser, error)

// NewGzip 表示支持 gzip 格式的压缩
func NewGzip(w io.Writer) (io.WriteCloser, error) {
	return gzip.NewWriter(w), nil
}

// NewDeflate 表示支持 deflate 压缩
func NewDeflate(w io.Writer) (io.WriteCloser, error) {
	return flate.NewWriter(w, flate.DefaultCompression)
}

type compress struct {
	h   http.Handler
	opt *Options
}

// New 构建一个支持压缩的中间件。
//
// 将 opt 传递给 New 之后，再修改 opt 中的值，将不再启作用。
func New(next http.Handler, opt *Options) http.Handler {
	opt.build()

	return &compress{
		h:   next,
		opt: opt,
	}
}

func (c *compress) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !c.opt.canCompressed(w) {
		c.h.ServeHTTP(w, r)
		return
	}

	accepts, err := accept.Parse(r.Header.Get("Accept-Encoding"))
	if err != nil {
		c.opt.ErrorLog.Println(err)
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	var gzw io.WriteCloser
	var accept *accept.Accept
	for _, accept = range accepts {
		// 不支持压缩
		if accept.Value == "identity" || accept.Value == "*" {
			break
		}

		f, found := c.opt.Funcs[accept.Value]
		if !found {
			continue
		}

		gzw, err = f(w)
		if err != nil { // 若出错，不压缩，直接返回
			c.opt.ErrorLog.Println(err)
			c.h.ServeHTTP(w, r)
			return
		}
		break // 找到需要的，则退出当前 for
	} // end for

	if gzw == nil { // 不支持的压缩格式
		c.h.ServeHTTP(w, r)
		return
	}

	w.Header().Set("Content-Encoding", accept.Value)
	w.Header().Add("Vary", "Accept-Encoding")
	resp := &response{
		gzw: gzw,
		rw:  w,
	}

	defer func() {
		// BUG(caixw):gzw.Close() 在执行时，可能会调用 w.Write() 进行输出，
		// 而 w.Write() 又有可能隐性调用 w.WriteHeader() 输出报头。
		//
		// 所以如果在 c.h 中进行 panic，并让外层接收后作报头输出，可能会出错，
		// 比如 issue9/web/internal/errors.Exit() 函数。
		if err := gzw.Close(); err != nil {
			c.opt.ErrorLog.Println(err)
		}
	}()

	// 此处可能 panic，所以得保证在 panic 之前，gzw 变量已经 Close
	c.h.ServeHTTP(resp, r)
}
