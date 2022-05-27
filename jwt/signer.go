// SPDX-License-Identifier: MIT

package jwt

import (
	"fmt"
	"io/fs"
	"math/rand"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/issue9/sliceutil"
	"github.com/issue9/web"
)

// Responser 向客户端输出的对象接口
type Responser interface {
	// SetAccessToken 设置令牌
	SetAccessToken(string)

	// SetRefreshToken 设置刷新令牌
	//
	// 未调用或是传递零值，输出时不应该带刷新令牌。
	SetRefreshToken(string)

	// SetExpires 设置过期时间
	//
	// 未调用或是传递零值，表示不需要输出时间信息。
	SetExpires(int)
}

type response struct {
	XMLName struct{} `json:"-" yaml:"-" xml:"token"`
	Access  string   `json:"access_token" yaml:"access_token" xml:"access_token"`
	Refresh string   `json:"refresh_token,omitempty" yaml:"refresh_token,omitempty" xml:"refresh_token,omitempty"`
	Expires int      `json:"expires,omitempty" yaml:"expires,omitempty" xml:"expires,attr,omitempty"`
}

// Signer 证书的签发管理
type Signer[T Claims] struct {
	keys    []*key
	expires int
	expired time.Duration
}

// NewResponse 默认的 Responser 接口实现
func NewResponse() Responser { return &response{} }

func NewSigner[T Claims](expired time.Duration) *Signer[T] {
	expires := int(expired.Seconds())
	return &Signer[T]{
		keys:    make([]*key, 0, 10),
		expires: expires,
		expired: expired,
	}
}

// RenderAccess 输出带令牌的对象
func (s *Signer[T]) RenderAccess(ctx *web.Context, status int, t Responser, accessClaims T) web.Responser {
	ac, err := s.Sign(accessClaims)
	if err != nil {
		return ctx.InternalServerError(err)
	}

	t.SetAccessToken(ac)
	t.SetExpires(s.expires)
	return ctx.Object(status, t)
}

// RenderAccessRefresh 输出带令牌和刷新令牌的对象
func (s *Signer[T]) RenderAccessRefresh(ctx *web.Context, status int, t Responser, accessClaims, refreshClaims T) web.Responser {
	ac, err := s.Sign(accessClaims)
	if err != nil {
		return ctx.InternalServerError(err)
	}

	rc, err := s.Sign(refreshClaims)
	if err != nil {
		return ctx.InternalServerError(err)
	}

	t.SetAccessToken(ac)
	t.SetRefreshToken(rc)
	t.SetExpires(s.expires)
	return ctx.Object(status, t)
}

// Sign 对 claims 进行签名
//
// 算法随机从 s.AddKey 添加的库里随机选取。
func (s *Signer[T]) Sign(claims T) (string, error) {
	var k *key
	switch l := len(s.keys); l {
	case 0:
		return "", ErrSigningMethodNotFound
	case 1:
		k = s.keys[0]
	default:
		k = s.keys[rand.Intn(l)]
	}

	t := jwt.NewWithClaims(k.sign, claims)
	t.Header["kid"] = k.id
	t.Header["alg"] = jwt.SigningMethodNone.Alg() // 不应该让用户知道算法，防止攻击。
	return t.SignedString(k.key)
}

func (s *Signer[T]) AddKey(id string, sign SigningMethod, private any) {
	if sliceutil.Exists(s.keys, func(e *key) bool { return e.id == id }) {
		panic(fmt.Sprintf("存在同名的签名方法 %s", id))
	}

	s.keys = append(s.keys, &key{
		id:   id,
		sign: sign,
		key:  private,
	})
}

func (s *Signer[T]) AddHMAC(id string, sign *jwt.SigningMethodHMAC, secret []byte) {
	s.AddKey(id, sign, secret)
}

func (s *Signer[T]) addRSA(id string, sign jwt.SigningMethod, private []byte) {
	pvt, err := jwt.ParseRSAPrivateKeyFromPEM(private)
	if err != nil {
		panic(err)
	}
	s.AddKey(id, sign, pvt)
}

func (s *Signer[T]) AddRSA(id string, sign *jwt.SigningMethodRSA, private []byte) {
	s.addRSA(id, sign, private)
}

func (s *Signer[T]) AddRSAPSS(id string, sign *jwt.SigningMethodRSAPSS, private []byte) {
	s.addRSA(id, sign, private)
}

func (s *Signer[T]) AddECDSA(id string, sign *jwt.SigningMethodECDSA, private []byte) {
	pvt, err := jwt.ParseECPrivateKeyFromPEM(private)
	if err != nil {
		panic(err)
	}
	s.AddKey(id, sign, pvt)
}

func (s *Signer[T]) AddEd25519(id string, sign *jwt.SigningMethodEd25519, private []byte) {
	pvt, err := jwt.ParseEdPrivateKeyFromPEM(private)
	if err != nil {
		panic(err)
	}

	s.AddKey(id, sign, pvt)
}

func (s *Signer[T]) AddRSAFromFS(id string, sign *jwt.SigningMethodRSA, fsys fs.FS, private string) {
	pvt, err := fs.ReadFile(fsys, private)
	if err != nil {
		panic(err)
	}
	s.AddRSA(id, sign, pvt)
}

func (s *Signer[T]) AddRSAPSSFromFS(id string, sign *jwt.SigningMethodRSAPSS, fsys fs.FS, private string) {
	pvt, err := fs.ReadFile(fsys, private)
	if err != nil {
		panic(err)
	}
	s.AddRSAPSS(id, sign, pvt)
}

func (s *Signer[T]) AddECDSAFromFS(id string, sign *jwt.SigningMethodECDSA, fsys fs.FS, private string) {
	pvt, err := fs.ReadFile(fsys, private)
	if err != nil {
		panic(err)
	}
	s.AddECDSA(id, sign, pvt)
}

func (s *Signer[T]) AddEd25519FromFS(id string, sign *jwt.SigningMethodEd25519, fsys fs.FS, private string) {
	pvt, err := fs.ReadFile(fsys, private)
	if err != nil {
		panic(err)
	}
	s.AddEd25519(id, sign, pvt)
}

func (resp *response) SetAccessToken(access string) { resp.Access = access }

func (resp *response) SetRefreshToken(refresh string) { resp.Refresh = refresh }

func (resp *response) SetExpires(expires int) { resp.Expires = expires }
