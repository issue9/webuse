// SPDX-License-Identifier: MIT

package jwt

import (
	"errors"
	"fmt"
	"io/fs"
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

type (
	// Verifier JWT 验证器
	Verifier[T Claims] struct {
		blocker       Blocker[T]
		keyFunc       jwt.Keyfunc
		claimsBuilder ClaimsBuilderFunc[T]
		keys          []*key
		keyIDs        []string
	}

	ClaimsBuilderFunc[T Claims] func() T

	contextKeyType int
)

// NewVerifier 声明 Verifier 对象
//
// d 为处理丢弃令牌的对象，如果为空表示不会对任何令牌作特殊处理；
// b 为 Claims 对象的生成方法；
func NewVerifier[T Claims](d Blocker[T], b ClaimsBuilderFunc[T]) *Verifier[T] {
	if d == nil {
		d = defaultBlocker[T]{}
	}

	j := &Verifier[T]{
		blocker:       d,
		claimsBuilder: b,
		keys:          make([]*key, 0, 10),
		keyIDs:        make([]string, 0, 10),
	}

	j.keyFunc = func(t *jwt.Token) (any, error) {
		if len(j.keys) == 0 || len(t.Header) == 0 {
			return nil, ErrSigningMethodNotFound
		}

		if kid, found := t.Header["kid"]; found {
			if k, found := sliceutil.At(j.keys, func(e *key) bool { return e.id == kid }); found {
				t.Method = k.sign // 忽略由用户指定的 header['alg']，而是有 kid 指定。
				return k.key, nil
			}
		}

		return nil, ErrSigningMethodNotFound
	}

	return j
}

// Middleware 解码用户的 token 并写入 *web.Context
func (j *Verifier[T]) Middleware(next web.HandlerFunc) web.HandlerFunc {
	return func(ctx *server.Context) web.Responser {
		h := j.GetToken(ctx)

		if j.blocker.TokenIsBlocked(h) {
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

		if j.blocker.ClaimsIsBlocked(t.Claims.(T)) {
			return ctx.Status(http.StatusUnauthorized)
		}

		ctx.Vars[contextKey] = t.Claims

		return next(ctx)
	}
}

// GetValue 返回解码后的 Claims 对象
func (j *Verifier[T]) GetValue(ctx *web.Context) (T, bool) {
	v, found := ctx.Vars[contextKey]
	if !found {
		var vv T
		return vv, false
	}
	return v.(T), true
}

func (j Verifier[T]) GetToken(ctx *web.Context) string {
	h := ctx.Request().Header.Get("Authorization")
	if len(h) > prefixLen && strings.ToLower(h[:prefixLen]) == prefix {
		h = h[prefixLen:]
	}
	return h
}

// KeyIDs 所有注册的编码名称
func (j *Verifier[T]) KeyIDs() []string { return j.keyIDs }

// AddKey 添加证书
func (j *Verifier[T]) AddKey(id string, sign SigningMethod, keyData any) {
	if sliceutil.Exists(j.keys, func(e *key) bool { return e.id == id }) {
		panic(fmt.Sprintf("存在同名的签名方法 %s", id))
	}

	j.keys = append(j.keys, &key{
		id:   id,
		sign: sign,
		key:  keyData,
	})
	j.keyIDs = append(j.keyIDs, id)
}

func (j *Verifier[T]) AddHMAC(id string, sign *jwt.SigningMethodHMAC, secret []byte) {
	j.AddKey(id, sign, secret)
}

func (j *Verifier[T]) addRSA(id string, sign jwt.SigningMethod, public []byte) {
	pub, err := jwt.ParseRSAPublicKeyFromPEM(public)
	if err != nil {
		panic(err)
	}
	j.AddKey(id, sign, pub)
}

func (j *Verifier[T]) AddRSA(id string, sign *jwt.SigningMethodRSA, public []byte) {
	j.addRSA(id, sign, public)
}

func (j *Verifier[T]) AddRSAPSS(id string, sign *jwt.SigningMethodRSAPSS, public []byte) {
	j.addRSA(id, sign, public)
}

func (j *Verifier[T]) AddECDSA(id string, sign *jwt.SigningMethodECDSA, public []byte) {
	pub, err := jwt.ParseECPublicKeyFromPEM(public)
	if err != nil {
		panic(err)
	}
	j.AddKey(id, sign, pub)
}

func (j *Verifier[T]) AddEd25519(id string, sign *jwt.SigningMethodEd25519, public []byte) {
	pub, err := jwt.ParseEdPublicKeyFromPEM(public)
	if err != nil {
		panic(err)
	}

	j.AddKey(id, sign, pub)
}

func (j *Verifier[T]) AddRSAFromFS(id string, sign *jwt.SigningMethodRSA, fsys fs.FS, public string) {
	pub, err := fs.ReadFile(fsys, public)
	if err != nil {
		panic(err)
	}
	j.AddRSA(id, sign, pub)
}

func (j *Verifier[T]) AddRSAPSSFromFS(id string, sign *jwt.SigningMethodRSAPSS, fsys fs.FS, public string) {
	pub, err := fs.ReadFile(fsys, public)
	if err != nil {
		panic(err)
	}
	j.AddRSAPSS(id, sign, pub)
}

func (j *Verifier[T]) AddECDSAFromFS(id string, sign *jwt.SigningMethodECDSA, fsys fs.FS, public string) {
	pub, err := fs.ReadFile(fsys, public)
	if err != nil {
		panic(err)
	}
	j.AddECDSA(id, sign, pub)
}

func (j *Verifier[T]) AddEd25519FromFS(id string, sign *jwt.SigningMethodEd25519, fsys fs.FS, public string) {
	pub, err := fs.ReadFile(fsys, public)
	if err != nil {
		panic(err)
	}
	j.AddEd25519(id, sign, pub)
}
