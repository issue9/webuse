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

// Manager 管理压缩处理函数的生成
type Manager struct {
	funcs map[string]WriterFunc

	size        int // 大于此值才会启用压缩
	prefix      []string
	vals        []string
	compressAll bool // 是否所有内容都压缩
}

// NewManager 声明新的 Manager 变量
func NewManager(funcs map[string]WriterFunc, mimetypes []string, size int) *Manager {
	prefix := make([]string, 0, len(mimetypes))
	vals := make([]string, 0, len(mimetypes))

	for _, typ := range mimetypes {
		if typ[len(typ)-1] == '*' {
			prefix = append(prefix, typ[:len(typ)-1])
		} else {
			vals = append(vals, typ)
		}
	}

	return &Manager{
		funcs:       funcs,
		size:        size,
		prefix:      prefix,
		vals:        vals,
		compressAll: len(funcs) > 0 && size == 0 && len(mimetypes) == 0,
	}
}

func (mgr *Manager) canCompressed(w http.ResponseWriter, errlog *log.Logger) bool {
	if mgr.compressAll {
		return true
	}

	if len(mgr.funcs) == 0 {
		return false
	}

	if mgr.size > 0 {
		l := w.Header().Get("Content-Length")
		if l != "" {
			ll, err := strconv.Atoi(l)
			if err != nil {
				errlog.Println(err)
				return false
			}

			if ll < mgr.size {
				return false
			}
		}
	}

	typ := w.Header().Get("Content-Type")
	if index := strings.IndexByte(typ, ';'); index > 0 {
		typ = strings.TrimSpace(typ[:index])
	}

	for _, val := range mgr.vals {
		if val == typ {
			return true
		}
	}

	for _, preifx := range mgr.prefix {
		if strings.HasPrefix(typ, preifx) {
			return true
		}
	}

	return false
}
