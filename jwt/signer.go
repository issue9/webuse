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

// Responser 向客户端输出令牌的数据结构
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

// Response 对 Responser 的默认实现
type Response struct {
	XMLName struct{} `json:"-" yaml:"-" xml:"token"`
	Access  string   `json:"access_token" yaml:"access_token" xml:"access_token"`
	Refresh string   `json:"refresh_token,omitempty" yaml:"refresh_token,omitempty" xml:"refresh_token,omitempty"`
	Expires int      `json:"expires,omitempty" yaml:"expires,omitempty" xml:"expires,attr,omitempty"`
}

// Signer 证书的签发管理
type Signer struct {
	keys    []*key
	expires int
	expired time.Duration

	refresh        bool
	refreshExpired time.Duration
}

// NewSigner 声明签名对象
//
// expired 普通令牌的过期时间；
// refresh 刷新令牌的时间，非零表示有刷新令牌，如果为非零值，则必须大于 expired；
func NewSigner(expired, refresh time.Duration) *Signer {
	if expired == 0 {
		panic("expired 必须大于 0")
	}

	if refresh != 0 && refresh <= expired {
		panic("refresh 必须大于 expired")
	}

	expires := int(expired.Seconds())
	return &Signer{
		keys: make([]*key, 0, 10),

		expires: expires,
		expired: expired,

		refresh:        refresh > 0,
		refreshExpired: refresh,
	}
}

// Render 输出带令牌的对象
func (s *Signer) Render(ctx *web.Context, status int, t Responser, accessClaims Claims) web.Responser {
	accessClaims.SetExpired(s.expired)
	ac, err := s.Sign(accessClaims)
	if err != nil {
		return ctx.InternalServerError(err)
	}
	t.SetAccessToken(ac)

	if s.refresh {
		r := accessClaims.BuildRefresh(ac)
		r.SetExpired(s.refreshExpired)
		rc, err := s.Sign(r)
		if err != nil {
			return ctx.InternalServerError(err)
		}
		t.SetRefreshToken(rc)
	}

	t.SetExpires(s.expires)
	return web.Object(status, t)
}

// Sign 对 claims 进行签名
//
// 算法随机从 s.AddKey 添加的库里随机选取。
func (s *Signer) Sign(claims Claims) (string, error) {
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

func (s *Signer) AddKey(id string, sign SigningMethod, private any) {
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
	s.AddKey(id, sign, secret)
}

func (s *Signer) addRSA(id string, sign jwt.SigningMethod, private []byte) {
	pvt, err := jwt.ParseRSAPrivateKeyFromPEM(private)
	if err != nil {
		panic(err)
	}
	s.AddKey(id, sign, pvt)
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
	s.AddKey(id, sign, pvt)
}

func (s *Signer) AddEd25519(id string, sign *jwt.SigningMethodEd25519, private []byte) {
	pvt, err := jwt.ParseEdPrivateKeyFromPEM(private)
	if err != nil {
		panic(err)
	}

	s.AddKey(id, sign, pvt)
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

func (resp *Response) SetAccessToken(access string) { resp.Access = access }

func (resp *Response) SetRefreshToken(refresh string) { resp.Refresh = refresh }

func (resp *Response) SetExpires(expires int) { resp.Expires = expires }
