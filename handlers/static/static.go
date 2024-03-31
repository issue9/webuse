// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package static 静态文件管理
package static

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/issue9/mux/v8/header"
	"github.com/issue9/web"
)

// AttachmentFileHandler 将 name 作为一个附件提供给客户端
//
// fsys 为文件系统；
// name 表示地址中表示文件名部分的参数名称；
// filename 为显示给客户端的文件名，如果为空，则会取 name 的文件名部分；
// inline 是否在浏览器内打开，主要看浏览器本身是否支持；
func AttachmentFileHandler(fsys fs.FS, name, filename string, inline bool) web.HandlerFunc {
	if fsys == nil {
		panic("参数 fsys 不能为空")
	}
	if name == "" {
		panic("参数 name 不能为空")
	}

	return func(ctx *web.Context) web.Responser {
		if p, found := ctx.Route().Params().Get(name); found { // 空值也是允许的值
			return AttachmentFile(ctx, fsys, p, filename, inline)
		}
		return ctx.NotFound()
	}
}

func AttachmentReaderHandler(filename string, inline bool, modtime time.Time, content io.ReadSeeker) web.HandlerFunc {
	return func(ctx *web.Context) web.Responser {
		return AttachmentReader(ctx, filename, inline, modtime, content)
	}
}

// ServeFileHandler 构建静态文件服务对象
//
// fsys 为文件系统；
// name 表示地址中表示文件名部分的参数名称。
// 以 val 表示 name 指定的参数，依以下规则读取内容：
//   - val 以 / 结尾，读取 val + index 指向的文件，若不存在返回 404;
//   - val 不以 / 结尾，则被当作普通的文件读取，若不存在返回 404，如果实际表示目录，则读取目录结构；
func ServeFileHandler(fsys fs.FS, name, index string) web.HandlerFunc {
	return func(ctx *web.Context) web.Responser {
		p, _ := ctx.Route().Params().Get(name) // 空值也是允许的值
		if p == "" {
			p = "."
		}

		stat, err := fs.Stat(fsys, p)
		if err != nil {
			return ctx.Error(err, "")
		}

		if !stat.IsDir() {
			http.ServeFileFS(ctx, ctx.Request(), fsys, p)
			return nil
		}

		index = path.Join(p, index)
		if stat, err = fs.Stat(fsys, index); err == nil && !stat.IsDir() {
			http.ServeFileFS(ctx, ctx.Request(), fsys, index)
			return nil
		} else if errors.Is(err, fs.ErrNotExist) || !stat.IsDir() {
			http.ServeFileFS(ctx, ctx.Request(), fsys, p) // 没找到 index 指向的文件，返回上一层的目录结构 p
			return nil
		} else {
			return ctx.Error(err, "")
		}
	}
}

// AttachmentFile 将文件作为下载对象
//
// name 为相对于 fsys 的文件名；
// filename 为显示给客户端的文件名，如果为空，则会取 name 的文件名部分；
// inline 是否在浏览器内打开，主要看浏览器本身是否支持；
func AttachmentFile(ctx *web.Context, fsys fs.FS, name, filename string, inline bool) web.Responser {
	if filename == "" {
		filename = filepath.Base(name)
	} else if strings.ContainsFunc(filename, func(r rune) bool { return r == '/' || r == filepath.Separator }) {
		panic(fmt.Sprintf("filename: %s 不能包含路径分隔符", filename))
	}
	filename = url.QueryEscape(filename) // 防止中文乱码

	attach := "attachment"
	if inline {
		attach = "inline"
	}

	cd := mime.FormatMediaType(attach, map[string]string{"filename": filename})
	ctx.Header().Set(header.ContentDisposition, cd)
	http.ServeFileFS(ctx, ctx.Request(), fsys, name)
	return nil
}

// AttachmentReader 将 [io.ReadSeeker] 作为下载对象
//
// modtime 表示展示给客户端的修改时间；
func AttachmentReader(ctx *web.Context, filename string, inline bool, modtime time.Time, content io.ReadSeeker) web.Responser {
	if strings.ContainsFunc(filename, func(r rune) bool { return r == '/' || r == filepath.Separator }) {
		panic(fmt.Sprintf("filename: %s 不能包含路径分隔符", filename))
	}

	attach := "attachment"
	if inline {
		attach = "inline"
	}

	cd := mime.FormatMediaType(attach, map[string]string{"filename": url.QueryEscape(filename)})
	ctx.Header().Set(header.ContentDisposition, cd)
	http.ServeContent(ctx, ctx.Request(), filename, modtime, content)
	return nil
}
