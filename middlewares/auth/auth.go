// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

// Package auth 登录凭证的验证
package auth

import "github.com/issue9/web"

// Auth 登录凭证的验证接口
type Auth interface {
	web.Middleware

	// Logout 退出
	Logout(*web.Context) error
}
