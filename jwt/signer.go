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

// BuildResponseFunc 根据给定的参数返回给定客户端的对象
//
// access 是必须的，表示请求数据的 token；
// refresh 表示刷新的 token，如果为空，不会输出；
// expires 表示 access 的过期时间；
type BuildResponseFunc func(access, refresh string, expires int) any

type Response struct {
	XMLName struct{} `json:"-" xml:"token"`
	Access  string   `json:"access_token" xml:"access_token"`
	Refresh string   `json:"refresh_token,omitempty" xml:"refresh_token,omitempty"`
	Expires int      `json:"expires,omitempty" xml:"expires,attr,omitempty"`
}

// Signer 证书的签发管理
//
// 仅负责对令牌的签发，如果需要验证令牌，则需要 [Verifier] 对象，
// 同时需要保证 [Verifier] 添加的证书数量和 ID 与当前对象是相同的。
type Signer struct {
	keys    []*key
	expires int
	expired time.Duration

	refresh        bool
	refreshExpired time.Duration

	br BuildResponseFunc
}

// NewSigner 声明签名对象
//
// expired 普通令牌的过期时间；
// refresh 刷新令牌的时间，非零表示有刷新令牌，如果为非零值，则必须大于 expired；
// br 表示将令牌组合成一个对象用以返回给客户端，可以为空，采用返回 [Response] 作为其默认实现；
func NewSigner(expired, refresh time.Duration, br BuildResponseFunc) *Signer {
	if expired == 0 {
		panic("expired 必须大于 0")
	}

	if refresh != 0 && refresh <= expired {
		panic("refresh 必须大于 expired")
	}

	if br == nil {
		br = func(access, refresh string, expires int) any {
			return &Response{Access: access, Refresh: refresh, Expires: expires}
		}
	}

	return &Signer{
		keys: make([]*key, 0, 10),

		expires: int(expired.Seconds()),
		expired: expired,

		refresh:        refresh > 0,
		refreshExpired: refresh,

		br: br,
	}
}

// Render 向客户端输出令牌
//
// 当前方法会将 accessClaims 进行签名，并返回 [web.Responser] 对象。
func (s *Signer) Render(ctx *web.Context, status int, accessClaims Claims) web.Responser {
	accessClaims.SetExpired(s.expired)
	ac, err := s.Sign(accessClaims)
	if err != nil {
		return ctx.InternalServerError(err)
	}

	var rc string
	if s.refresh {
		r := accessClaims.BuildRefresh(ac)
		r.SetExpired(s.refreshExpired)
		rc, err = s.Sign(r)
		if err != nil {
			return ctx.InternalServerError(err)
		}
	}

	return web.Object(status, s.br(ac, rc, s.expires))
}

// Sign 对 claims 进行签名
//
// 算法随机从 [Signer.AddKey] 添加的库里选取。
func (s *Signer) Sign(claims Claims) (string, error) {
	var k *key
	switch l := len(s.keys); l {
	case 0:
		return "", ErrSigningMethodNotFound()
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

func (s *Signer) addKey(id string, sign SigningMethod, private any) {
	if sliceutil.Exists(s.keys, func(e *key) bool { return e.id == id }) {
		panic(fmt.Sprintf("存在同名的签名方法 %s", id))
	}

	s.keys = append(s.keys, &key{
		id:   id,
		sign: sign,
		key:  private,
	})
}

func (s *Signer) AddHMAC(id string, sign *jwt.SigningMethodHMAC, secret []byte) {
	s.addKey(id, sign, secret)
}

func (s *Signer) addRSA(id string, sign SigningMethod, private []byte) {
	pvt, err := jwt.ParseRSAPrivateKeyFromPEM(private)
	if err != nil {
		panic(err)
	}
	s.addKey(id, sign, pvt)
}

func (s *Signer) AddRSA(id string, sign *jwt.SigningMethodRSA, private []byte) {
	s.addRSA(id, sign, private)
}

func (s *Signer) AddRSAPSS(id string, sign *jwt.SigningMethodRSAPSS, private []byte) {
	s.addRSA(id, sign, private)
}

func (s *Signer) AddECDSA(id string, sign *jwt.SigningMethodECDSA, private []byte) {
	pvt, err := jwt.ParseECPrivateKeyFromPEM(private)
	if err != nil {
		panic(err)
	}
	s.addKey(id, sign, pvt)
}

func (s *Signer) AddEd25519(id string, sign *jwt.SigningMethodEd25519, private []byte) {
	pvt, err := jwt.ParseEdPrivateKeyFromPEM(private)
	if err != nil {
		panic(err)
	}

	s.addKey(id, sign, pvt)
}

func (s *Signer) AddRSAFromFS(id string, sign *jwt.SigningMethodRSA, fsys fs.FS, private string) {
	pvt, err := fs.ReadFile(fsys, private)
	if err != nil {
		panic(err)
	}
	s.AddRSA(id, sign, pvt)
}

func (s *Signer) AddRSAPSSFromFS(id string, sign *jwt.SigningMethodRSAPSS, fsys fs.FS, private string) {
	pvt, err := fs.ReadFile(fsys, private)
	if err != nil {
		panic(err)
	}
	s.AddRSAPSS(id, sign, pvt)
}

func (s *Signer) AddECDSAFromFS(id string, sign *jwt.SigningMethodECDSA, fsys fs.FS, private string) {
	pvt, err := fs.ReadFile(fsys, private)
	if err != nil {
		panic(err)
	}
	s.AddECDSA(id, sign, pvt)
}

func (s *Signer) AddEd25519FromFS(id string, sign *jwt.SigningMethodEd25519, fsys fs.FS, private string) {
	pvt, err := fs.ReadFile(fsys, private)
	if err != nil {
		panic(err)
	}
	s.AddEd25519(id, sign, pvt)
}

// Add 添加签名方法
func (s *Signer) Add(id string, sign SigningMethod, private []byte) {
	switch m := sign.(type) {
	case *jwt.SigningMethodHMAC:
		s.AddHMAC(id, m, private)
	case *jwt.SigningMethodRSA:
		s.AddRSA(id, m, private)
	case *jwt.SigningMethodRSAPSS:
		s.AddRSAPSS(id, m, private)
	case *jwt.SigningMethodECDSA:
		s.AddECDSA(id, m, private)
	case *jwt.SigningMethodEd25519:
		s.AddEd25519(id, m, private)
	default:
		panic(invalidSignForID(id))
	}
}

// AddFromFS 添加签名方法密钥从文件中加载
func (s *Signer) AddFromFS(id string, sign SigningMethod, fsys fs.FS, private string) {
	switch m := sign.(type) {
	case *jwt.SigningMethodRSA:
		s.AddRSAFromFS(id, m, fsys, private)
	case *jwt.SigningMethodRSAPSS:
		s.AddRSAPSSFromFS(id, m, fsys, private)
	case *jwt.SigningMethodECDSA:
		s.AddECDSAFromFS(id, m, fsys, private)
	case *jwt.SigningMethodEd25519:
		s.AddEd25519FromFS(id, m, fsys, private)
	default:
		panic(invalidSignForID(id))
	}
}
