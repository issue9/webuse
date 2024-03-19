// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims Claims 对象需要实现的接口
type Claims interface {
	jwt.Claims

	// BuildRefresh 根据令牌 token 生成刷新令牌的 [Claims]
	//
	// token 生成的刷新令牌与此令牌相关联，此值可通过返回对象的 [Claims.BaseToken] 返回；
	// created 表示刷新令牌的刷新时间，一般为 [time.Now]；
	BuildRefresh(token string, created time.Time) Claims

	// BaseToken 刷新令牌关联的令牌
	//
	// 如果是非刷新令牌的话，此方法应该返回空字符串。
	// 否则应该返回调用 [Claims.BuildRefresh] 的参数。
	BaseToken() string
}
