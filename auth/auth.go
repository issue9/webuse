// SPDX-License-Identifier: MIT

// Package auth 验证类的中间件
package auth

type keyType int

// ValueKey 保存于 context 中的值的名称
//
// 所有的验证中间件，在验证成功之后，都会将一个值附加在 r.Context()
// 之上，可以通过 ValueKey 获取其相应的值。
const ValueKey keyType = 0
