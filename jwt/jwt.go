// SPDX-License-Identifier: MIT

// Package jwt JSON Web Tokens 验证
//
//  sign := NewSigner[*jwt.RegisterClaims](...)
//  v := NewVerifier[*jwt.RegisterClaims](nil, builder)
//
//  // 添加多种编码方式
//  sign.Add("hmac", jwt.SigningMethodHS256, []byte("secret"))
//  v.Add("hmac", jwt.SigningMethodHS256, []byte("secret"))
//  sign.AddRSA("rsa", jwt.SigningMethodRS256, []byte("private"))
//  v.AddRSA("rsa", jwt.SigningMethodRS256, []byte("public"))
//
//  sign.Sign(&jwt.RegisterClaims{...})
//  sign.Sign(&jwt.RegisterClaims{...})
package jwt

import (
	"errors"

	"github.com/golang-jwt/jwt/v4"
)

var ErrSigningMethodNotFound = errors.New("jwt: 算法未找到")

type (
	Claims = jwt.Claims

	SigningMethod = jwt.SigningMethod

	key struct {
		id   any
		sign SigningMethod
		key  any // 公钥或是私钥
	}
)
