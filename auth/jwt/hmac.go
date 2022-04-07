// SPDX-License-Identifier: MIT

package jwt

import (
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/issue9/web"
	"github.com/issue9/web/server"

	"github.com/issue9/middleware/v6/auth"
)

type hmac struct {
	jwt    *JWT
	sign   *jwt.SigningMethodHMAC
	secret []byte
}

func (j *JWT) NewHMAC(secret []byte, sign *jwt.SigningMethodHMAC) Middleware {
	return &hmac{
		jwt:    j,
		sign:   sign,
		secret: []byte(secret),
	}
}

// Sign 生成 token
func (h *hmac) Sign(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(h.sign, claims)
	return token.SignedString(h.secret)
}

// Middleware 解码用户的 token
//
// 可通过 auth.GetValue 获取解码后的值。
func (h *hmac) Middleware(next web.HandlerFunc) web.HandlerFunc {
	return func(ctx *server.Context) web.Responser {
		token := ctx.Request().Header.Get("Authorization")
		t, err := jwt.ParseWithClaims(token, h.jwt.claimsBuilder(), func(token *jwt.Token) (interface{}, error) {
			return h.secret, nil
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
