// SPDX-License-Identifier: MIT

package jwt

import (
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/issue9/web"
	"github.com/issue9/web/server"

	"github.com/issue9/middleware/v6/auth"
)

type hmacMiddleware struct {
	jwt    *JWT
	sign   *jwt.SigningMethodHMAC
	secret []byte
}

func (j *JWT) NewHMAC(secret []byte, sign *jwt.SigningMethodHMAC) Middleware {
	return &hmacMiddleware{
		jwt:    j,
		sign:   sign,
		secret: []byte(secret),
	}
}

// Sign 生成 token
func (m *hmacMiddleware) Sign(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(m.sign, claims)
	return token.SignedString(m.secret)
}

// Middleware 解码用户的 token
func (m *hmacMiddleware) Middleware(next web.HandlerFunc) web.HandlerFunc {
	return func(ctx *server.Context) web.Responser {
		token := ctx.Request().Header.Get("Authorization")
		t, err := jwt.ParseWithClaims(token, m.jwt.claimsBuilder(), func(token *jwt.Token) (interface{}, error) {
			return m.secret, nil
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
