// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

// Package swagger 提供 [SwaggerUI] 的实现
//
// [SwaggerUI]: https://swagger.io
package swagger

import (
	"embed"

	"github.com/issue9/web"
	"github.com/issue9/web/mimetype/html"
	"github.com/issue9/web/openapi"

	"github.com/issue9/webuse/v7/plugins/openapi/internal"
)

//go:embed *.html
var tpl embed.FS

// OnlineAssets swagger 的 CDN 资源
const OnlineAssets = "https://unpkg.com/swagger-ui-dist@5.18.2"

// Install 安装模板
func Install(s web.Server) {
	html.Install(s, internal.Funcs, "*.html", tpl)
}

// WithHTML 指定 [swagger] 的 HTML 模板
//
// 可用来代替 [openapi.WithHTML]
//
// assets swagger 的页面资源，可以直接引用 [OnlineAssets]，也可以指向自有的服务器；
// logo 图标；
//
// [swagger]: https://swagger.io/docs/open-source-tools/swagger-ui/usage/installation/
func WithHTML(assets, logo string) openapi.Option {
	return openapi.WithHTML("swagger", assets, logo)
}

// WithOnlineHTML 采用 [OnlineAssets] 作为参数的 [WithHTML] 版本
//
// 如果 favicon 为空，则会采用 [OnlineAssets] 下的默认图标。
func WithOnlineHTML(favicon string) openapi.Option {
	if favicon == "" {
		favicon = OnlineAssets + "/favicon-32x32.png"
	}
	return WithHTML(OnlineAssets, favicon)
}
