// SPDX-FileCopyrightText: 2024-2025 caixw
//
// SPDX-License-Identifier: MIT

// package openapi 提供 OpenAPI 的文档阅读器
package openapi

import (
	"fmt"

	"github.com/issue9/web"
	"github.com/issue9/web/openapi"

	"github.com/issue9/webuse/v7/plugins/openapi/scalar"
	"github.com/issue9/webuse/v7/plugins/openapi/swagger"
)

// Viewer 创建 OpenAPI 文档阅读器
//
// name 指定阅读器的名称，当前可用的值：
//   - scalar https://github.com/scalar/scalar
//   - swagger https://swagger.io
//
// logo 文档的 LOGO，可以为空，根据 name 值的不同，
// 对空值的处理会有所不同，具体可参考各自的 WithHTML 文档；
func Viewer(s web.Server, name, logo string) openapi.Option {
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
