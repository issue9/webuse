// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

// Package auth 登录凭证的验证
package auth

import (
	"strings"

	"github.com/issue9/web"
)

const Bearer = "bearer "

// Auth 登录凭证的验证接口
//
// T 表示每次验证后，附加在 [web.Context] 上的数据。
type Auth[T any] interface {
	web.Middleware

	// Logout 退出
	Logout(*web.Context) error

	// GetInfo 获取用户数据
	//
	// 当验证通过之后，验证接口同时会将用户信息写入到 [web.Context]
	// 可通过当前方法获取写入的数据。
	GetInfo(*web.Context) (T, bool)
}

// GetToken 获取客户端提交的令牌
//
// header 表示报头的名称；
// prefix 表示报头内容的前缀；
func GetToken(ctx *web.Context, prefix, header string) string {
	h := ctx.Request().Header.Get(header)
	if l := len(prefix); len(h) > l && strings.ToLower(h[:l]) == prefix {
		return h[l:]
	}
	ctx.Logs().DEBUG().LocaleString(web.Phrase("the client %s header %s is invalid format", header, h))
	return ""
}

// BuildToken 生成一个完整的令牌
func BuildToken(prefix, token string) string { return prefix + token }
