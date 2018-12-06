// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package compress

import (
	"log"
	"net/http"
	"strconv"
	"strings"
)

// Options New 的参数
type Options struct {
	Funcs    map[string]WriterFunc
	Size     int // 大于此值才会启用压缩
	ErrorLog *log.Logger

	// 仅对该表中的类型进行压缩
	//
	// 可以用通配符 * 作为结尾：
	// text* 表示以 text 开头的 mimetype；
	// 其它类型的值，表示完全匹配。
	Types []string

	// Types 列表的处理结果保存在 prefixTypes 和 types 中。
	//
	// prefixTypes 保存通配符匹配的值列表；
	// types 表示完全匹配的值列表。
	prefixTypes []string
	types       []string
}

func (opt *Options) build() {
	prefix := make([]string, 0, len(opt.Types))
	types := make([]string, 0, len(opt.Types))

	for _, typ := range opt.Types {
		if typ[len(typ)-1] == '*' {
			prefix = append(prefix, typ[:len(typ)-1])
		} else {
			types = append(types, typ)
		}
	}

	opt.prefixTypes = prefix
	opt.types = types
}

func (opt *Options) canCompressed(w http.ResponseWriter) bool {
	if len(opt.Funcs) == 0 {
		return false
	}

	if opt.Size > 0 {
		l := w.Header().Get("Content-Length")
		if l != "" {
			ll, err := strconv.Atoi(l)
			if err != nil {
				opt.ErrorLog.Println(err)
				return false
			}

			if ll < opt.Size {
				return false
			}
		}
	}

	typ := w.Header().Get("Content-Type")
	if index := strings.IndexByte(typ, ';'); index > 0 {
		typ = strings.TrimSpace(typ[:index])
	}

	for _, val := range opt.types {
		if val == typ {
			return true
		}
	}

	for _, preifx := range opt.prefixTypes {
		if strings.HasPrefix(typ, preifx) {
			return true
		}
	}

	return false
}
