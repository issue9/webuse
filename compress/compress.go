// SPDX-License-Identifier: MIT

// Package compress 提供一个支持内容压缩的中间件
package compress

import (
	"log"
	"net/http"
	"strings"

	"github.com/issue9/sliceutil"
)

// Compress 提供压缩功能的中件间
//
// NOTE: Compress 必须是所有有输出功能中间件的最外层。
// 否则可能造成部分内容被压缩，而部分内容未压缩的情况。
type Compress struct {
	// 如果指定了这个值，那么会把错误日志输出到此。
	// 若未指定，则不输出内容。
	errlog *log.Logger

	algorithms []*algorithm // 按添加顺序保存，查找 * 时按添加顺序进行比对。

	typePrefix []string // 保存通配符匹配的值列表；
	types      []string // 表示完全匹配的值列表。
	anyTypes   bool     // 表示任何类型都需要压缩

	ignoreMethods []string
}

// Default 简单的初始化 Compress 方式
//
// ignoreMethods 被设置为  HEAD 和 OPTIONS；同时添加 deflate, gzip 和 br 三种压缩方式。
func Default(errlog *log.Logger, types ...string) *Compress {
	chk := func(ok bool) {
		if !ok {
			panic("存在相同的算法名称")
		}
	}

	c := New(errlog, []string{http.MethodHead, http.MethodOptions}, types...)
	chk(c.AddAlgorithm("deflate", NewDeflate))
	chk(c.AddAlgorithm("gzip", NewGzip))
	chk(c.AddAlgorithm("br", NewBrotli))

	return c
}

// New 构建一个支持压缩的中间件
//
// errlog 错误日志的输出通道；
// ignoreMethods 忽略的请求方法，如果不为空，则这些请求方法的请求将不会被压缩；
// types 表示需要进行压缩处理的 mimetype 类型，可以是以下格式：
//  - application/json 具体类型；
//  - text* 表示以 text 开头的所有类型；
//  - * 表示所有类型，一旦指定此值，则其它设置都将被忽略；
func New(errlog *log.Logger, ignoreMethods []string, types ...string) *Compress {
	if errlog == nil {
		panic("参数 errlog 不能为空")
	}

	c := &Compress{
		algorithms:    make([]*algorithm, 0, 4),
		errlog:        errlog,
		ignoreMethods: ignoreMethods,
	}

	c.typePrefix = make([]string, 0, len(types))
	c.types = make([]string, 0, len(types))
	c.AddType(types...)

	return c
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
			c.anyTypes = true
		case typ[len(typ)-1] == '*':
			c.typePrefix = append(c.typePrefix, typ[:len(typ)-1])
		default:
			c.types = append(c.types, typ)
		}
	}
}

// DeleteType 删除对媒体类型的支持
//
// types 的格式可参考 AddType 方法。
//
// NOTE: 仅用于删除通过 AddType 添加的内容。
func (c *Compress) DeleteType(types ...string) {
	for _, typ := range types {
		switch {
		case typ == "*":
			c.anyTypes = false
			c.typePrefix = c.typePrefix[:0]
			c.types = c.types[:0]
		case typ[len(typ)-1] == '*':
			index := sliceutil.Delete(c.typePrefix, func(i int) bool { return strings.HasPrefix(c.typePrefix[i], typ[:len(typ)-1]) })
			c.typePrefix = c.typePrefix[:index]

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
		if len(c.algorithms) == 0 || c.isIgnore(r.Method) {
			next.ServeHTTP(w, r)
			return
		}

		name, wf, notAcceptable := c.findAlgorithm(r)
		if notAcceptable {
			w.WriteHeader(http.StatusNotAcceptable)
			return
		}

		if wf == nil {
			next.ServeHTTP(w, r)
			return
		}

		resp := c.newResponse(w, wf, name)
		defer resp.close()
		next.ServeHTTP(resp, r) // 此处可能 panic，所以得保证在 panic 之前，resp 已经关闭
	})
}

func (c *Compress) isIgnore(method string) bool {
	for _, m := range c.ignoreMethods {
		if m == method {
			return true
		}
	}
	return false
}

func (c *Compress) canCompressed(typ string) bool {
	if c.anyTypes {
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

	for _, prefix := range c.typePrefix {
		if strings.HasPrefix(typ, prefix) {
			return true
		}
	}

	return false
}
