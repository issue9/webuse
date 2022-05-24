// SPDX-License-Identifier: MIT

// Package jwt JSON Web Tokens 验证
package jwt

import (
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/issue9/sliceutil"
	"github.com/issue9/web"
	"github.com/issue9/web/server"
)

const (
	contextKey contextKeyType = 1

	prefix = "bearer "

	prefixLen = 7 // len(prefix)
)

var ErrKidNotFound = errors.New("jwt: 指定的算法未找到")

type (
	Claims        = jwt.Claims
	SigningMethod = jwt.SigningMethod

	ClaimsBuilderFunc[T Claims] func() T

	// JWT 生成 JWT 令牌管理
	//
	// 可以指定多个证书，如果存在多个证书，那么将通过 header["kid"] 查询每个令牌对应的证书，
	// 如果未指定 kid，那么始终会拿第一个添加的证书作为令牌的证书。
	JWT[T Claims] struct {
		discarder     Discarder[T]
		keyFunc       jwt.Keyfunc
		claimsBuilder ClaimsBuilderFunc[T]
		keys          []*key
	}

	contextKeyType int
)

// New 声明 JWT 对象
//
// d 为处理丢弃令牌的对象，如果为空表示不会对任何令牌作特殊处理；
// b 为 Claims 对象的生成方法；
func New[T Claims](d Discarder[T], b ClaimsBuilderFunc[T]) *JWT[T] {
	if d == nil {
		d = defaultDiscarder[T]{}
	}

	j := &JWT[T]{
		discarder:     d,
		claimsBuilder: b,
	}
	j.keyFunc = func(t *jwt.Token) (any, error) {
		if k := j.findKey(t.Header); k != nil {
			return k.public, nil
		}
		return nil, ErrKidNotFound
	}

	return j
}

func (j *JWT[T]) findKey(headers map[string]any) *key {
	kid, found := headers["kid"]
	if !found {
		return j.keys[0]
	}

	k, _ := sliceutil.At(j.keys, func(e *key) bool { return e.id == kid })
	return k
}

// Sign 生成 token
//
// headers 表示输出的 JWT.Header 中的字段，通过 headers["kid"] 可指定算法；
func (j *JWT[T]) Sign(claims Claims, headers map[string]any) (string, error) {
	if k := j.findKey(headers); k != nil {
		t := jwt.NewWithClaims(k.sign, claims)
		for k, v := range headers {
			t.Header[k] = v
		}
		return t.SignedString(k.private)
	}
	return "", ErrKidNotFound
}

// Middleware 解码用户的 token 并写入 *web.Context
func (j *JWT[T]) Middleware(next web.HandlerFunc) web.HandlerFunc {
	return func(ctx *server.Context) web.Responser {
		h := j.GetToken(ctx)

		if j.discarder.TokenIsDiscarded(h) {
			return ctx.Status(http.StatusUnauthorized)
		}

		t, err := jwt.ParseWithClaims(h, j.claimsBuilder(), j.keyFunc)

		if errors.Is(err, &jwt.ValidationError{}) {
			ctx.Logs().ERROR().Error(err)
			return ctx.Status(http.StatusUnauthorized)
		} else if err != nil {
			return ctx.InternalServerError(err)
		}

		if !t.Valid {
			return ctx.Status(http.StatusUnauthorized)
		}

		if j.discarder.ClaimsIsDiscarded(t.Claims.(T)) {
			return ctx.Status(http.StatusUnauthorized)
		}

		ctx.Vars[contextKey] = t.Claims

		return next(ctx)
	}
}

// GetValue 返回解码后的 Claims 对象
func (j *JWT[T]) GetValue(ctx *web.Context) (T, bool) {
	v, found := ctx.Vars[contextKey]
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
