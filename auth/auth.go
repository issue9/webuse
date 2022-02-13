// SPDX-License-Identifier: MIT

// Package auth 验证类的中间件
package auth

type keyType int

// ValueKey 保存于 web.Context 中的值的名称
const ValueKey keyType = 0
