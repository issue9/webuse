// SPDX-License-Identifier: MIT

package compress

import (
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"

	"github.com/andybalholm/brotli"
	"github.com/issue9/qheader"
	"github.com/issue9/sliceutil"
)

// Writer 所有压缩对象实现的接口
type Writer interface {
	io.WriteCloser
	Reset(io.Writer)
}

// WriterFunc 定义了将一个 io.Writer 声明为具有压缩功能的 io.WriteCloser
type WriterFunc func(w io.Writer) (Writer, error)

type algorithm struct {
	name   string
	writer WriterFunc
}

// NewGzip 新建 gzip 算法
func NewGzip(w io.Writer) (Writer, error) {
	return gzip.NewWriter(w), nil
}

// NewDeflate 新建 deflate 算法
func NewDeflate(w io.Writer) (Writer, error) {
	return flate.NewWriter(w, flate.DefaultCompression)
}

// NewBrotli 新建 br 算法
func NewBrotli(w io.Writer) (Writer, error) {
	return brotli.NewWriter(w), nil
}

// AddAlgorithm 添加压缩算法
//
// 如果已经存在，返回 true。
// 当前用户的 Accept-Encoding 的匹配到 * 时，按添加顺序查找真正的匹配项。
// 不能添加名为 identity 和 * 的算法。
//
// 如果未添加任何算法，则每个请求都相当于是 identity 规则。
func (c *Compress) AddAlgorithm(name string, w WriterFunc) (exists bool) {
	if name == "" || name == "identity" || name == "*" {
		panic("name 值不能为 identity 和 *")
	}

	if w == nil {
		panic("参数 w 不能为空")
	}

	if sliceutil.Count(c.algorithms, func(i int) bool { return c.algorithms[i].name == name }) > 0 {
		return true
	}

	c.algorithms = append(c.algorithms, algorithm{name: name, writer: w})
	return false
}

// SetAlgorithm 设置压缩算法
//
// 如果 w 为 nil，则表示去掉此算法的支持。
func (c *Compress) SetAlgorithm(name string, w WriterFunc) {
	if name == "" || name == "identity" || name == "*" {
		panic("name 值不能为 identity 和 *")
	}

	if w == nil {
		size := sliceutil.Delete(c.algorithms, func(i int) bool { return c.algorithms[i].name == name })
		c.algorithms = c.algorithms[:size]
		return
	}

	c.algorithms = append(c.algorithms, algorithm{name: name, writer: w})
}

// 如果返回的 f 为空值，表示不需要压缩
func (c *Compress) findAlgorithm(r *http.Request) (name string, f WriterFunc, notAcceptable bool) {
	accepts := qheader.AcceptEncoding(r)
	for _, accept := range accepts {
		if accept.Err != nil {
			c.printError(accept.Err)
			continue
		}

		if accept.Value == "*" {
			if accept.Q == 0.0 {
				return "", nil, true
			}

			for _, a := range c.algorithms {
				for _, item := range accepts {
					if item.Value != a.name {
						return a.name, a.writer, false
					}
				}
			}
			continue
		}

		if accept.Value == "identity" {
			if accept.Q == 0.0 {
				return "", nil, true
			}
			return "", nil, false
		}

		for _, a := range c.algorithms {
			if a.name == accept.Value {
				return a.name, a.writer, false
			}
		}
	}

	return // 没有匹配，表示不需要进行压缩
}
