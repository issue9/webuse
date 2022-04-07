// SPDX-License-Identifier: MIT

package jwt

import (
	"crypto/rsa"
	"io/fs"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v4"
	"github.com/issue9/web"
	"github.com/issue9/web/server"

	"github.com/issue9/middleware/v6/auth"
)

type rsaMiddleware struct {
	jwt     *JWT
	sign    *jwt.SigningMethodRSA
	private *rsa.PrivateKey
	public  *rsa.PublicKey
}

func (j *JWT) NewRSAFromFile(prvKey, pubKey string, sign *jwt.SigningMethodRSA) (Middleware, error) {
	prv, err := os.ReadFile(prvKey)
	if err != nil {
		return nil, err
	}

	pub, err := os.ReadFile(pubKey)
	if err != nil {
		return nil, err
	}

	return j.NewRSA(prv, pub, sign)
}

func (j *JWT) NewRSAFromFS(fsys fs.FS, prvKey, pubKey string, sign *jwt.SigningMethodRSA) (Middleware, error) {
	prv, err := fs.ReadFile(fsys, prvKey)
	if err != nil {
		return nil, err
	}

	pub, err := fs.ReadFile(fsys, pubKey)
	if err != nil {
		return nil, err
	}

	return j.NewRSA(prv, pub, sign)
}

func (j *JWT) NewRSA(prvKey, pubKey []byte, sign *jwt.SigningMethodRSA) (Middleware, error) {
	prv, err := jwt.ParseRSAPrivateKeyFromPEM(prvKey)
	if err != nil {
		return nil, err
	}

	pub, err := jwt.ParseRSAPublicKeyFromPEM(pubKey)
	if err != nil {
		return nil, err
	}

	return &rsaMiddleware{
		jwt:     j,
		sign:    sign,
		private: prv,
		public:  pub,
	}, nil
}

func (m *rsaMiddleware) Sign(c jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(m.sign, c)
	return token.SignedString(m.private)
}

func (m *rsaMiddleware) Middleware(next web.HandlerFunc) web.HandlerFunc {
	return func(ctx *server.Context) web.Responser {
		token := ctx.Request().Header.Get("Authorization")
		t, err := jwt.ParseWithClaims(token, m.jwt.claimsBuilder(), func(token *jwt.Token) (interface{}, error) {
			return m.public, nil
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
