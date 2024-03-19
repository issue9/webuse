// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package jwt

import "github.com/golang-jwt/jwt/v5"

// Claims Claims 对象需要实现的接口
type Claims interface {
	jwt.Claims

	// BuildRefresh 根据令牌 token 生成刷新令牌的 [Claims]
	//
	// 返回对象也是 [Claims] 的实现，可以通过返回对象的 [Claims.BaseToken] 获得关联的基础令牌。
	BuildRefresh(string) Claims

	// BaseToken 刷新令牌关联的令牌
	//
	// 如果是非刷新令牌的话，此方法应该返回空字符串。
	// 否则应该返回调用 [Claims.BuildRefresh] 的参数。
	BaseToken() string
}
