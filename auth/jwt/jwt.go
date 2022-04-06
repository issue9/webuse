// SPDX-License-Identifier: MIT

// Package jwt JWT 验证
package jwt

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/issue9/web"
	"github.com/issue9/web/server"

	"github.com/issue9/middleware/v6/auth"
)

type ClaimsBuilderFunc func() jwt.Claims

type JWT struct {
	claimsBuilder ClaimsBuilderFunc
	secret        []byte
	maxAge        time.Duration
	sign          jwt.SigningMethod
}

func New(b ClaimsBuilderFunc, secret string, maxAge time.Duration, sign jwt.SigningMethod) *JWT {
	return &JWT{
		claimsBuilder: b,
		secret:        []byte(secret),
		maxAge:        maxAge,
		sign:          sign,
	}
}

// Sign 生成 token
func (j *JWT) Sign(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(j.sign, claims)
	return token.SignedString(j.secret)
}

// Middleware 解码用户的 token
//
// 可通过 auth.GetValue 获取解码后的值。
func (j *JWT) Middleware(next web.HandlerFunc) web.HandlerFunc {
	return func(ctx *server.Context) web.Responser {
		token := ctx.Request().Header.Get("Authorization")
		t, err := jwt.ParseWithClaims(token, j.claimsBuilder(), func(token *jwt.Token) (interface{}, error) {
			return j.secret, nil
		})

		if err != nil {
			return ctx.InternalServerError(err)
		}

		if !t.Valid {
			return ctx.Status(http.StatusUnauthorized)
		}

		auth.SetValue(ctx, t.Claims)

		return next(ctx)
	}
}
