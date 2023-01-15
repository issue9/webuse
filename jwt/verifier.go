// SPDX-License-Identifier: MIT

package jwt

import (
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/issue9/sliceutil"
	"github.com/issue9/web"
)

const (
	contextKey contextKeyType = 1

	prefix = "bearer "

	prefixLen = 7 // len(prefix)
)

type (
	// Verifier JWT 验证器
	//
	// 仅负责对令牌的验证，如果需要签发令牌，则需要 [Signer] 对象，
	// 同时需要保证 [Signer] 添加的证书数量和 ID 与当前对象是相同的。
	Verifier[T Claims] struct {
		blocker       Blocker[T]
		keyFunc       jwt.Keyfunc
		claimsBuilder BuildClaimsFunc[T]
		keys          []*key
	}

	BuildClaimsFunc[T Claims] func() T

	contextKeyType int
)

// NewVerifier 声明 Verifier 对象
//
// b 为处理丢弃令牌的对象，如果为空表示不会对任何令牌作特殊处理；
// f 为 Claims 对象的生成方法；
func NewVerifier[T Claims](b Blocker[T], f BuildClaimsFunc[T]) *Verifier[T] {
	if b == nil {
		b = defaultBlocker[T]{}
	}

	j := &Verifier[T]{
		blocker:       b,
		claimsBuilder: f,
		keys:          make([]*key, 0, 10),
	}

	j.keyFunc = func(t *jwt.Token) (any, error) {
		if len(j.keys) == 0 || len(t.Header) == 0 {
			return nil, ErrSigningMethodNotFound()
		}

		if kid, found := t.Header["kid"]; found {
			if k, found := sliceutil.At(j.keys, func(e *key) bool { return e.id == kid }); found {
				t.Method = k.sign // 忽略由用户指定的 header['alg']，而是有 kid 指定。
				return k.key, nil
			}
		}

		return nil, ErrSigningMethodNotFound()
	}

	return j
}

// Middleware 解码用户的 token 并写入 *web.Context
func (j *Verifier[T]) Middleware(next web.HandlerFunc) web.HandlerFunc {
	return func(ctx *web.Context) web.Responser {
		h := j.GetToken(ctx)
		if h == "" || j.blocker.TokenIsBlocked(h) {
			return ctx.Problem("401")
		}

		t, err := jwt.ParseWithClaims(h, j.claimsBuilder(), j.keyFunc)
		if errors.Is(err, &jwt.ValidationError{}) {
			ctx.Logs().ERROR().Error(err)
			return ctx.Problem(web.ProblemUnauthorized)
		} else if err != nil {
			return ctx.InternalServerError(err)
		}

		if !t.Valid {
			return ctx.Problem(web.ProblemUnauthorized)
		}

		if j.blocker.ClaimsIsBlocked(t.Claims.(T)) {
			return ctx.Problem(web.ProblemUnauthorized)
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

// GetToken 获取客户端提交的 token
func (j Verifier[T]) GetToken(ctx *web.Context) string {
	h := ctx.Request().Header.Get("Authorization")
	if len(h) > prefixLen && strings.ToLower(h[:prefixLen]) == prefix {
		h = h[prefixLen:]
	}
	return h
}

func (j *Verifier[T]) addKey(id string, sign SigningMethod, keyData any) {
	if sliceutil.Exists(j.keys, func(e *key) bool { return e.id == id }) {
		panic(fmt.Sprintf("存在同名的签名方法 %s", id))
	}

	j.keys = append(j.keys, &key{
		id:   id,
		sign: sign,
		key:  keyData,
	})
}

func (j *Verifier[T]) AddHMAC(id string, sign *jwt.SigningMethodHMAC, secret []byte) {
	j.addKey(id, sign, secret)
}

func (j *Verifier[T]) addRSA(id string, sign jwt.SigningMethod, public []byte) {
	pub, err := jwt.ParseRSAPublicKeyFromPEM(public)
	if err != nil {
		panic(err)
	}
	j.addKey(id, sign, pub)
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
	j.addKey(id, sign, pub)
}

func (j *Verifier[T]) AddEd25519(id string, sign *jwt.SigningMethodEd25519, public []byte) {
	pub, err := jwt.ParseEdPublicKeyFromPEM(public)
	if err != nil {
		panic(err)
	}

	j.addKey(id, sign, pub)
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

// Add 添加签名方法
func (j *Verifier[T]) Add(id string, sign jwt.SigningMethod, public []byte) {
	switch m := sign.(type) {
	case *jwt.SigningMethodHMAC:
		j.AddHMAC(id, m, public)
	case *jwt.SigningMethodRSA:
		j.AddRSA(id, m, public)
	case *jwt.SigningMethodRSAPSS:
		j.AddRSAPSS(id, m, public)
	case *jwt.SigningMethodECDSA:
		j.AddECDSA(id, m, public)
	case *jwt.SigningMethodEd25519:
		j.AddEd25519(id, m, public)
	default:
		panic(invalidSignForID(id))
	}
}

// AddFromFS 添加签名方法密钥从文件中加载
func (j *Verifier[T]) AddFromFS(id string, sign jwt.SigningMethod, fsys fs.FS, public string) {
	switch m := sign.(type) {
	case *jwt.SigningMethodRSA:
		j.AddRSAFromFS(id, m, fsys, public)
	case *jwt.SigningMethodRSAPSS:
		j.AddRSAPSSFromFS(id, m, fsys, public)
	case *jwt.SigningMethodECDSA:
		j.AddECDSAFromFS(id, m, fsys, public)
	case *jwt.SigningMethodEd25519:
		j.AddEd25519FromFS(id, m, fsys, public)
	default:
		panic(invalidSignForID(id))
	}
}

func invalidSignForID(id string) string { return fmt.Sprintf("%s 对应的签名方法无效", id) }
