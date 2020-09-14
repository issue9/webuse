// SPDX-License-Identifier: MIT

// Package compress 提供一个支持内容压缩的中间件
package compress

import (
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"

	"github.com/andybalholm/brotli"
	"github.com/issue9/qheader"
)

// WriterFunc 定义了将一个 io.Writer 声明为具有压缩功能的 io.WriteCloser
type WriterFunc func(w io.Writer) (io.WriteCloser, error)

// NewGzip 新建 gzip 算法
func NewGzip(w io.Writer) (io.WriteCloser, error) {
	return gzip.NewWriter(w), nil
}

// NewDeflate 新建 deflate 算法
func NewDeflate(w io.Writer) (io.WriteCloser, error) {
	return flate.NewWriter(w, flate.DefaultCompression)
}

// NewBrotli 新建 br 算法
func NewBrotli(w io.Writer) (io.WriteCloser, error) {
	return brotli.NewWriter(w), nil
}

// Compress 提供压缩功能的中件间
type Compress struct {
	h   http.Handler
	opt *Options
}

// New 构建一个支持压缩的中间件
//
// 将 opt 传递给 New 之后，再修改 opt 中的值，将不再启作用。
func New(next http.Handler, opt *Options) *Compress {
	if opt != nil {
		opt.build()
	}

	return &Compress{
		h:   next,
		opt: opt,
	}
}

func (c *Compress) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if c.opt == nil || len(c.opt.Funcs) == 0 {
		c.h.ServeHTTP(w, r)
		return
	}

	accepts, err := qheader.AcceptEncoding(r)
	if err != nil {
		c.opt.println(err)
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	var wf WriterFunc
	var accept *qheader.Header
	for _, accept = range accepts {
		if accept.Value == "identity" || accept.Value == "*" { // 不支持压缩
			break
		}

		if f, found := c.opt.Funcs[accept.Value]; found {
			wf = f
			break
		}
	} // end for

	if wf == nil { // 客户端不需要压缩
		c.h.ServeHTTP(w, r)
		return
	}

	resp := &response{
		rw:           w,
		opt:          c.opt,
		f:            wf,
		encodingName: accept.Value,
	}

	defer resp.close()

	// 此处可能 panic，所以得保证在 panic 之前，resp 已经关闭
	c.h.ServeHTTP(resp, r)
}
