// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

// Package locales 加载本地化的插件
package locales

import (
	"io/fs"

	"github.com/issue9/web"
)

// Locales 本地化内容管理
type Locales struct {
	glob string
	f    []fs.FS
}

// New 声明 [Locales]
//
// glob 表示从 f 中查找文件的模式；
// f 表示存放本地化文件的文件系统；
func New(glob string, f ...fs.FS) *Locales {
	return &Locales{
		glob: glob,
		f:    f,
	}
}

// Append 添加新的文件系统
func (l *Locales) Append(f ...fs.FS) *Locales {
	l.f = append(l.f, f...)
	return l
}

func (l *Locales) Plugin(s web.Server) { s.Locale().LoadMessages(l.glob, l.f...) }
