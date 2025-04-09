// SPDX-FileCopyrightText: 2025 caixw
//
// SPDX-License-Identifier: MIT

// Package scalar 提供了 [scalar] 的实现
//
// [scalar]: https://github.com/scalar/scalar
package scalar

import (
	"embed"

	"github.com/issue9/web"
	"github.com/issue9/web/mimetype/html"
	"github.com/issue9/web/openapi"

	"github.com/issue9/webuse/v7/internal/openapifuncs"
)

//go:embed *.html
var tpl embed.FS

// CDNAssets scalar 的 CDN 资源
const CDNAssets = "https://cdn.jsdelivr.net/npm/@scalar/api-reference@1.28.14"

// Install 安装模板
//
// NOTE: 此操作会同时安装 json 和 yaml 两个模板函数
func Install(s web.Server) { html.Install(s, openapifuncs.Funcs, nil, "*.html", tpl) }

// WithHTML 指定 [scalar] 的 HTML 模板
//
// 可用来代替 [openapi.WithHTML]
//
// assets scalar 的页面资源，可以直接引用 [CDNAssets]，
// 或是采用 [files] 与 [static] 搭建一个本地的静态文件服务；
//
// [scalar]: https://github.com/scalar/scalar/blob/main/documentation/configuration.md
// [static]: https://github.com/issue9/webuse/handlers/static
// [files]: https://github.com/swaggo/files
func WithHTML(assets, logo string) openapi.Option {
	return openapi.WithHTML("scalar", assets, logo)
}

// WithCDN 采用 [CDNAssets] 作为参数的 [WithHTML] 版本
func WithCDN(logo string) openapi.Option { return WithHTML(CDNAssets, logo) }
