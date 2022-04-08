// SPDX-License-Identifier: MIT

// Package jwt JWT 验证
package jwt

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/issue9/web"
	"github.com/issue9/web/server"

	"github.com/issue9/middleware/v6/auth"
)

const prefix = "bearer "

const prefixLen = 7 // len(prefix)

type (
	ClaimsBuilderFunc func() jwt.Claims

	JWT struct {
		claimsBuilder   ClaimsBuilderFunc
		signFunc        jwt.SigningMethod
		private, public any
	}
)

func New(b ClaimsBuilderFunc, signFunc jwt.SigningMethod, private, public any) *JWT {
	return &JWT{
		claimsBuilder: b,
		signFunc:      signFunc,
		private:       private,
		public:        public,
	}
}

func NewHMAC(b ClaimsBuilderFunc, signFunc *jwt.SigningMethodHMAC, secret []byte) *JWT {
	return New(b, signFunc, secret, secret)
}

func NewRSA(b ClaimsBuilderFunc, sign *jwt.SigningMethodRSA, private, public []byte) (*JWT, error) {
	return newRSA(b, sign, private, public)
}

func NewRSAPSS(b ClaimsBuilderFunc, sign *jwt.SigningMethodRSAPSS, private, public []byte) (*JWT, error) {
	return newRSA(b, sign, private, public)
}

func newRSA(b ClaimsBuilderFunc, sign jwt.SigningMethod, private, public []byte) (*JWT, error) {
	pvt, err := jwt.ParseRSAPrivateKeyFromPEM(private)
	if err != nil {
		println("pvt:", err.Error())
		return nil, err
	}

	pub, err := jwt.ParseRSAPublicKeyFromPEM(public)
	if err != nil {
		println("pub:", err.Error(), string(public))
		return nil, err
	}

	return New(b, sign, pvt, pub), nil
}

func NewECDSA(b ClaimsBuilderFunc, sign *jwt.SigningMethodECDSA, private, public []byte) (*JWT, error) {
	pvt, err := jwt.ParseECPrivateKeyFromPEM(private)
	if err != nil {
		return nil, err
	}

	pub, err := jwt.ParseECPublicKeyFromPEM(public)
	if err != nil {
		return nil, err
	}

	return New(b, sign, pvt, pub), nil
}

func NewEd25519(b ClaimsBuilderFunc, sign *jwt.SigningMethodEd25519, private, public []byte) (*JWT, error) {
	pvt, err := jwt.ParseEdPrivateKeyFromPEM(private)
	if err != nil {
		return nil, err
	}

	pub, err := jwt.ParseEdPublicKeyFromPEM(public)
	if err != nil {
		return nil, err
	}

	return New(b, sign, pvt, pub), nil
}

// Sign 生成 token
func (j *JWT) Sign(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(j.signFunc, claims)
	return token.SignedString(j.private)
}

// Middleware 解码用户的 token 并写入 *web.Context
//
// 如果需要提交，可以采用 auth.GetValue 函数。
func (j *JWT) Middleware(next web.HandlerFunc) web.HandlerFunc {
	return func(ctx *server.Context) web.Responser {
		h := ctx.Request().Header.Get("Authorization")
		if len(h) > prefixLen && strings.ToLower(h[:prefixLen]) == prefix {
			h = h[prefixLen:]
		}

		t, err := jwt.ParseWithClaims(h, j.claimsBuilder(), func(token *jwt.Token) (interface{}, error) {
			return j.public, nil
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
