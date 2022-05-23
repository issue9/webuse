// SPDX-License-Identifier: MIT

// Package jwt JSON Web Tokens 验证
package jwt

import (
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/issue9/web"
	"github.com/issue9/web/server"
)

type keyType int

const valueKey keyType = 1

const prefix = "bearer "

const prefixLen = 7 // len(prefix)

type (
	ClaimsBuilderFunc[T jwt.Claims] func() T

	JWT[T jwt.Claims] struct {
		discarder       Discarder
		claimsBuilder   ClaimsBuilderFunc[T]
		signFunc        jwt.SigningMethod
		private, public any
	}

	// Discarder 判断令牌是否被丢弃
	//
	// 在某些情况下，需要强制用户的令牌不再可用，可以使用 Discarder 接口，
	// 当 JWT 接受此对象时，将采用 IsDiscarded 来判断令牌是否是被丢弃的。
	Discarder interface {
		// IsDiscarded 令牌是否已被提早丢弃
		IsDiscarded(string) bool
	}

	defaultDiscarder struct{}
)

func (d defaultDiscarder) IsDiscarded(string) bool { return true }

// New 声明 JWT 对象
//
// b 为 Claims 对象的生成方法；
// private 和 public 为公私钥数据，如果是 hmac 算法，则两者是一样的值；
func New[T jwt.Claims](d Discarder, b ClaimsBuilderFunc[T], signFunc jwt.SigningMethod, private, public any) *JWT[T] {
	if d == nil {
		d = defaultDiscarder{}
	}

	return &JWT[T]{
		discarder:     d,
		claimsBuilder: b,
		signFunc:      signFunc,
		private:       private,
		public:        public,
	}
}

func NewHMAC[T jwt.Claims](d Discarder, b ClaimsBuilderFunc[T], sign *jwt.SigningMethodHMAC, secret []byte) *JWT[T] {
	return New(d, b, sign, secret, secret)
}

func NewRSA[T jwt.Claims](d Discarder, b ClaimsBuilderFunc[T], sign *jwt.SigningMethodRSA, private, public []byte) (*JWT[T], error) {
	return newRSA(d, b, sign, private, public)
}

func NewRSAPSS[T jwt.Claims](d Discarder, b ClaimsBuilderFunc[T], sign *jwt.SigningMethodRSAPSS, private, public []byte) (*JWT[T], error) {
	return newRSA(d, b, sign, private, public)
}

func newRSA[T jwt.Claims](d Discarder, b ClaimsBuilderFunc[T], sign jwt.SigningMethod, private, public []byte) (*JWT[T], error) {
	pvt, err := jwt.ParseRSAPrivateKeyFromPEM(private)
	if err != nil {
		return nil, err
	}

	pub, err := jwt.ParseRSAPublicKeyFromPEM(public)
	if err != nil {
		return nil, err
	}

	return New(d, b, sign, pvt, pub), nil
}

func NewECDSA[T jwt.Claims](d Discarder, b ClaimsBuilderFunc[T], sign *jwt.SigningMethodECDSA, private, public []byte) (*JWT[T], error) {
	pvt, err := jwt.ParseECPrivateKeyFromPEM(private)
	if err != nil {
		return nil, err
	}

	pub, err := jwt.ParseECPublicKeyFromPEM(public)
	if err != nil {
		return nil, err
	}

	return New(d, b, sign, pvt, pub), nil
}

func NewEd25519[T jwt.Claims](d Discarder, b ClaimsBuilderFunc[T], sign *jwt.SigningMethodEd25519, private, public []byte) (*JWT[T], error) {
	pvt, err := jwt.ParseEdPrivateKeyFromPEM(private)
	if err != nil {
		return nil, err
	}

	pub, err := jwt.ParseEdPublicKeyFromPEM(public)
	if err != nil {
		return nil, err
	}

	return New(d, b, sign, pvt, pub), nil
}

// Sign 生成 token
func (j *JWT[T]) Sign(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(j.signFunc, claims)
	return token.SignedString(j.private)
}

// Middleware 解码用户的 token 并写入 *web.Context
func (j *JWT[T]) Middleware(next web.HandlerFunc) web.HandlerFunc {
	return func(ctx *server.Context) web.Responser {
		h := j.GetToken(ctx)

		if j.discarder.IsDiscarded(h) {
			return ctx.Status(http.StatusUnauthorized)
		}

		t, err := jwt.ParseWithClaims(h, j.claimsBuilder(), func(token *jwt.Token) (interface{}, error) {
			return j.public, nil
		})

		if errors.Is(err, &jwt.ValidationError{}) {
			ctx.Logs().ERROR().Error(err)
			return ctx.Status(http.StatusUnauthorized)
		} else if err != nil {
			return ctx.InternalServerError(err)
		}

		if !t.Valid {
			return ctx.Status(http.StatusUnauthorized)
		}
		ctx.Vars[valueKey] = t.Claims

		return next(ctx)
	}
}

// GetValue 返回解码后的  Claims 对象
func (j *JWT[T]) GetValue(ctx *web.Context) (T, bool) {
	v, found := ctx.Vars[valueKey]
	if !found {
		var vv T
		return vv, false
	}
	return v.(T), true
}

func (j JWT[T]) GetToken(ctx *web.Context) string {
	h := ctx.Request().Header.Get("Authorization")
	if len(h) > prefixLen && strings.ToLower(h[:prefixLen]) == prefix {
		h = h[prefixLen:]
	}
	return h
}
