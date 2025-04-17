// SPDX-FileCopyrightText: 2024-2025 caixw
//
// SPDX-License-Identifier: MIT

// package openapis 为 [https://pkg.go.dev/github.com/issue9/web/openapi] 提供功能支持
package openapis

import (
	"fmt"

	"github.com/issue9/web"
	"github.com/issue9/web/openapi"

	"github.com/issue9/webuse/v7/openapis/scalar"
	"github.com/issue9/webuse/v7/openapis/swagger"
)

// WithCDNViewer 创建基于 CDN 的 OpenAPI 文档阅读器
//
// name 指定阅读器的名称，当前可用的值：
//   - scalar https://github.com/scalar/scalar
//   - swagger https://swagger.io
//
// logo 文档的 LOGO，可以为空，根据 name 值的不同，
// 对空值的处理会有所不同，具体可参考各自的 WithCDN 文档；
//
// NOTE: 若需要更精细的控制，可以直接调用 [github.com/issue9/webuse/v7/openapis/scalar]、
// [github.com/issue9/webuse/v7/openapis/swagger] 包中的方法。
func WithCDNViewer(s web.Server, name, logo string) openapi.Option {
	switch name {
	case "swagger":
		swagger.Install(s)
		return swagger.WithCDN(logo)
	case "scalar":
		scalar.Install(s)
		return scalar.WithCDN(logo)
	default:
		panic(fmt.Sprintf("参数 name 的值 %s 不是一个有效的值", name))
	}
}
