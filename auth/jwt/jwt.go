// SPDX-License-Identifier: MIT

// Package jwt JWT 验证
package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/issue9/web"
)

type (
	ClaimsBuilderFunc func() jwt.Claims

	// Middleware JWT 中间件需要实现的接口
	Middleware interface {
		// Sign 返回对 claims 加密后的数据
		Sign(jwt.Claims) (string, error)

		// Middleware 将解码后的 Claims 写入 *web.Context
		//
		// NOTE: 可通过 auth.GetValue 获取解码后的值。
		Middleware(web.HandlerFunc) web.HandlerFunc
	}

	JWT struct {
		claimsBuilder ClaimsBuilderFunc
		maxAge        time.Duration
	}
)

func New(b ClaimsBuilderFunc, maxAge time.Duration) *JWT {
	return &JWT{
		claimsBuilder: b,
		maxAge:        maxAge,
	}
}
