// SPDX-License-Identifier: MIT

// Package compress 提供一个支持内容压缩的中间件
package compress

import (
	"compress/flate"
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/andybalholm/brotli"
	"github.com/issue9/qheader"
	"github.com/issue9/sliceutil"
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
	// 指定压缩名称对应的生成函数
	writers map[string]WriterFunc

	// 如果指定了这个值，那么会把错误日志输出到此。
	// 若未指定，则不输出内容。
	errlog *log.Logger

	any bool // 表示任何类型都需要压缩

	// Types 列表的处理结果保存在 prefixTypes 和 types 中
	//
	// prefix 保存通配符匹配的值列表；
	// types 表示完全匹配的值列表。
	prefix []string
	types  []string
}

// New 构建一个支持压缩的中间件
//
// types 表示需要进行压缩处理的 mimetype 类型，可以是以下格式：
//  - application/json 具体类型；
//  - text* 表示以 text 开头的所有类型；
//  - * 表示所有类型，一旦指定此值，则其它设置都将被忽略；
func New(errlog *log.Logger, writers map[string]WriterFunc, types ...string) *Compress {
	c := &Compress{
		writers: make(map[string]WriterFunc, len(writers)),
		errlog:  errlog,
	}

	for name, w := range writers {
		c.SetWriter(name, w)
	}

	c.prefix = make([]string, 0, len(types))
	c.types = make([]string, 0, len(types))
	c.AddType(types...)

	return c
}

// SetWriter 设置压缩算法
//
// 如果 w 为 nil，则表示去掉此算法的支持。
func (c *Compress) SetWriter(name string, w WriterFunc) {
	if w == nil {
		delete(c.writers, name)
		return
	}

	c.writers[name] = w
}

// AddType 添加对媒体类型的支持
//
// types 表示需要进行压缩处理的 mimetype 类型，可以是以下格式：
//  - application/json 具体类型；
//  - text* 表示以 text 开头的所有类型；
//  - * 表示所有类型，一旦指定此值，则其它设置都将被忽略；
func (c *Compress) AddType(types ...string) {
	for _, typ := range types {
		switch {
		case typ == "*":
			c.any = true
		case typ[len(typ)-1] == '*':
			c.prefix = append(c.prefix, typ[:len(typ)-1])
		default:
			c.types = append(c.types, typ)
		}
	}
}

// DeleteType 删除对媒体类型的支持
//
// types 的格式可参考 AddType 方法。
func (c *Compress) DeleteType(types ...string) {
	for _, typ := range types {
		switch {
		case typ == "*":
			c.any = false
			c.prefix = c.prefix[:0]
			c.types = c.types[:0]
		case typ[len(typ)-1] == '*':
			index := sliceutil.Delete(c.prefix, func(i int) bool { return strings.HasPrefix(c.prefix[i], typ[:len(typ)-1]) })
			c.prefix = c.prefix[:index]

			index = sliceutil.Delete(c.types, func(i int) bool { return strings.HasPrefix(c.types[i], typ[:len(typ)-1]) })
			c.types = c.types[:index]
		default:
			index := sliceutil.Delete(c.types, func(i int) bool { return c.types[i] == typ })
			c.types = c.types[:index]
		}
	}
}

// MiddlewareFunc 将当前中间件应用于 next
func (c *Compress) MiddlewareFunc(next func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return c.Middleware(http.HandlerFunc(next))
}

// Middleware 将当前中间件应用于 next
func (c *Compress) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(c.writers) == 0 {
			next.ServeHTTP(w, r)
			return
		}

		accepts := qheader.AcceptEncoding(r)
		var wf WriterFunc
		var accept *qheader.Header
		for _, accept = range accepts {
			if accept.Err != nil {
				c.printError(accept.Err)
				continue
			}

			if accept.Value == "identity" || accept.Value == "*" { // 不支持压缩
				break
			}

			if f, found := c.writers[accept.Value]; found { // 找到压缩函数
				wf = f
				break
			}
		}

		if wf == nil || accept == nil { // 客户端不需要压缩
			next.ServeHTTP(w, r)
			return
		}

		resp := &response{
			rw:           w,
			c:            c,
			f:            wf,
			encodingName: accept.Value,
		}

		defer resp.close()

		// 此处可能 panic，所以得保证在 panic 之前，resp 已经关闭
		next.ServeHTTP(resp, r)
	})
}

func (c *Compress) canCompressed(typ string) bool {
	if len(c.writers) == 0 {
		return false
	}

	if c.any {
		return true
	}

	if index := strings.IndexByte(typ, ';'); index > 0 {
		typ = strings.TrimSpace(typ[:index])
	}

	for _, val := range c.types {
		if val == typ {
			return true
		}
	}

	for _, prefix := range c.prefix {
		if strings.HasPrefix(typ, prefix) {
			return true
		}
	}

	return false
}

func (c *Compress) printError(err error) {
	if c.errlog != nil {
		c.errlog.Println(err)
	}
}
