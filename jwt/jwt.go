// SPDX-License-Identifier: MIT

// Package jwt JSON Web Tokens 验证
//
//	sign := NewSigner(...)
//	v := NewVerifier[*jwt.RegisterClaims](nil, builder)
//
//	// 添加多种编码方式
//	sign.Add("hmac", jwt.SigningMethodHS256, []byte("secret"))
//	v.Add("hmac", jwt.SigningMethodHS256, []byte("secret"))
//	sign.AddRSA("rsa", jwt.SigningMethodRS256, []byte("private"))
//	v.AddRSA("rsa", jwt.SigningMethodRS256, []byte("public"))
//
//	sign.Sign(&jwt.RegisterClaims{...})
//	sign.Sign(&jwt.RegisterClaims{...})
package jwt

import (
	"errors"
	"io/fs"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/issue9/web"
)

var errSigningMethodNotFound = errors.New("jwt: 算法未找到")

type (
	// Claims JWT Claims 对象需要实现的接口
	Claims interface {
		jwt.Claims

		// BuildRefresh 根据令牌生成刷新令牌用的 Claims
		BuildRefresh(string) Claims

		// SetExpired 设置过期时间
		SetExpired(time.Duration)
	}

	SigningMethod = jwt.SigningMethod

	key struct {
		id   any
		sign SigningMethod
		key  any // 公钥或是私钥
	}

	// JWT JWT 管理
	//
	// 同时包含了 [Verifier] 和 [Signer]，大部分时候这是比直接使用 [Verifier] 和 [Signer] 更方便的方法。
	JWT[T Claims] struct {
		v *Verifier[T]
		s *Signer
	}
)

func ErrSigningMethodNotFound() error { return errSigningMethodNotFound }

func New[T Claims](b Blocker[T], f BuildClaimsFunc[T], expired, refresh time.Duration, br BuildResponseFunc) *JWT[T] {
	v := NewVerifier(b, f)
	s := NewSigner(expired, refresh, br)
	return &JWT[T]{v: v, s: s}
}

// Middleware 解码用户的 token 并写入 *web.Context
func (j *JWT[T]) Middleware(next web.HandlerFunc) web.HandlerFunc { return j.v.Middleware(next) }

// GetValue 返回解码后的 Claims 对象
func (j *JWT[T]) GetValue(ctx *web.Context) (T, bool) { return j.v.GetValue(ctx) }

// GetToken 获取客户端提交的 token
func (j JWT[T]) GetToken(ctx *web.Context) string { return j.v.GetToken(ctx) }

// Render 向客户端输出令牌
//
// 当前方法会将 accessClaims 进行签名，并返回 [web.Responser] 对象。
func (j *JWT[T]) Render(ctx *web.Context, status int, accessClaims Claims) web.Responser {
	return j.s.Render(ctx, status, accessClaims)
}

// Sign 对 claims 进行签名
//
// 算法随机从 [Signer.AddKey] 添加的库里选取。
func (j *JWT[T]) Sign(claims Claims) (string, error) { return j.s.Sign(claims) }

// AddHMAC 添加 HMAC 算法
//
// NOTE: 调用者需要保证每次重启之后，id 值不能改变，否则所有的登录信息 token 将失效。
func (j *JWT[T]) AddHMAC(id string, sign *jwt.SigningMethodHMAC, secret []byte) {
	j.v.addKey(id, sign, secret)
	j.s.addKey(id, sign, secret)
}

func (j *JWT[T]) AddRSA(id string, sign *jwt.SigningMethodRSA, pub, pvt []byte) {
	j.v.AddRSA(id, sign, pub)
	j.s.AddRSA(id, sign, pvt)
}

func (j *JWT[T]) AddRSAPSS(id string, sign *jwt.SigningMethodRSAPSS, pub, pvt []byte) {
	j.v.AddRSAPSS(id, sign, pub)
	j.s.AddRSAPSS(id, sign, pvt)
}

func (j *JWT[T]) AddECDSA(id string, sign *jwt.SigningMethodECDSA, pub, pvt []byte) {
	j.v.AddECDSA(id, sign, pub)
	j.s.AddECDSA(id, sign, pvt)

}

func (j *JWT[T]) AddEd25519(id string, sign *jwt.SigningMethodEd25519, pub, pvt []byte) {
	j.v.AddEd25519(id, sign, pub)
	j.s.AddEd25519(id, sign, pvt)
}

func (j *JWT[T]) AddRSAFromFS(id string, sign *jwt.SigningMethodRSA, fsys fs.FS, pub, pvt string) {
	j.v.AddRSAFromFS(id, sign, fsys, pub)
	j.s.AddRSAFromFS(id, sign, fsys, pvt)
}

func (j *JWT[T]) AddRSAPSSFromFS(id string, sign *jwt.SigningMethodRSAPSS, fsys fs.FS, pub, pvt string) {
	j.v.AddRSAPSSFromFS(id, sign, fsys, pub)
	j.s.AddRSAPSSFromFS(id, sign, fsys, pvt)
}

func (j *JWT[T]) AddECDSAFromFS(id string, sign *jwt.SigningMethodECDSA, fsys fs.FS, pub, pvt string) {
	j.v.AddECDSAFromFS(id, sign, fsys, pub)
	j.s.AddECDSAFromFS(id, sign, fsys, pvt)
}

func (j *JWT[T]) AddEd25519FromFS(id string, sign *jwt.SigningMethodEd25519, fsys fs.FS, pub, pvt string) {
	j.v.AddEd25519FromFS(id, sign, fsys, pub)
	j.s.AddEd25519FromFS(id, sign, fsys, pvt)
}

// Add 添加签名方法
//
// NOTE: 如果添加的是 HMAC 类型的函数，那么 pvt 和 pub 两者需要相同。
func (j *JWT[T]) Add(id string, sign jwt.SigningMethod, pub, pvt []byte) {
	j.v.Add(id, sign, pub)
	j.s.Add(id, sign, pvt)
}

// AddFromFS 添加签名方法密钥从文件中加载
//
// NOTE: 此方法不支持 HMAC 类型。
func (j *JWT[T]) AddFromFS(id string, sign jwt.SigningMethod, fsys fs.FS, pub, pvt string) {
	j.v.AddFromFS(id, sign, fsys, pub)
	j.s.AddFromFS(id, sign, fsys, pvt)
}
