// SPDX-FileCopyrightText: 2015-2024 caixw
//
// SPDX-License-Identifier: MIT

package jwt

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/issue9/assert/v4"
	"github.com/issue9/web"
	xjson "github.com/issue9/web/mimetype/json"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/servertest"

	"github.com/issue9/webuse/v7/internal/mauth"
	"github.com/issue9/webuse/v7/middlewares/auth"
)

var _ auth.Auth[*testClaims] = &JWT[*testClaims]{}

func newJWT(a *assert.Assertion, expired, refresh time.Duration) (web.Server, *JWT[*testClaims]) {
	s, err := server.New("test", "1.0.0", &server.Options{
		HTTPServer: &http.Server{Addr: ":8080"},
		Mimetypes:  server.JSONMimetypes(),
	})
	a.NotError(err).NotNil(s).
		NotError(s.Cache().Clean())

	m := NewCacheBlocker[*testClaims](s, "test_", expired, refresh)
	b := func() *testClaims { return &testClaims{} }
	j := New(m, b, expired, refresh, nil)
	a.NotNil(j)

	return s, j
}

func TestJWT_Middleware(t *testing.T) {
	a := assert.New(t, false)
	fsys := os.DirFS("./testdata")

	s, j := newJWT(a, time.Hour, time.Hour*2)
	j.Add("hmac-secret", jwt.SigningMethodHS256, []byte("secret"), []byte("secret"))
	verifierMiddleware(a, s, j)

	a.PanicString(func() {
		j.AddHMAC("hmac-secret", jwt.SigningMethodHS256, []byte("secret"))
	}, "存在同名的签名方法 hmac-secret")

	a.PanicString(func() {
		j.s.AddHMAC("hmac-secret", jwt.SigningMethodHS256, []byte("secret"))
	}, "存在同名的签名方法 hmac-secret")

	pub, pvt := readFile(a, fsys, "rsa-public.pem", "rsa-private.pem")
	s, j = newJWT(a, time.Hour, time.Hour*2)
	j.Add("rsa", jwt.SigningMethodRS256, pub, pvt)
	verifierMiddleware(a, s, j)
	s, j = newJWT(a, time.Hour, time.Hour*2)
	j.AddRSA("rsa-2", jwt.SigningMethodRS256, pub, pvt)
	verifierMiddleware(a, s, j)

	pub, pvt = readFile(a, fsys, "rsa-public.pem", "rsa-private.pem")
	s, j = newJWT(a, time.Hour, time.Hour*2)
	j.Add("rsa-pss", jwt.SigningMethodPS256, pub, pvt)
	verifierMiddleware(a, s, j)
	s, j = newJWT(a, time.Hour, time.Hour*2)
	j.AddRSAPSS("rsa-pss-2", jwt.SigningMethodPS256, pub, pvt)
	verifierMiddleware(a, s, j)

	pub, pvt = readFile(a, fsys, "ec256-public.pem", "ec256-private.pem")
	s, j = newJWT(a, time.Hour, time.Hour*2)
	j.Add("ecdsa", jwt.SigningMethodES256, pub, pvt)
	verifierMiddleware(a, s, j)
	s, j = newJWT(a, time.Hour, time.Hour*2)
	j.AddECDSA("ecdsa-2", jwt.SigningMethodES256, pub, pvt)
	verifierMiddleware(a, s, j)

	pub, pvt = readFile(a, fsys, "ed25519-public.pem", "ed25519-private.pem")
	s, j = newJWT(a, time.Hour, time.Hour*2)
	j.Add("ed25519", jwt.SigningMethodEdDSA, pub, pvt)
	verifierMiddleware(a, s, j)
	s, j = newJWT(a, time.Hour, time.Hour*2)
	j.AddEd25519("ed25519-2", jwt.SigningMethodEdDSA, pub, pvt)
	verifierMiddleware(a, s, j)

	// 一次性加载多个
	s, j = newJWT(a, time.Hour, time.Hour*2)
	j.AddFromFS("ed25519", jwt.SigningMethodEdDSA, fsys, "ed25519-public.pem", "ed25519-private.pem")
	j.AddFromFS("ecdsa", jwt.SigningMethodES256, fsys, "ec256-public.pem", "ec256-private.pem")
	j.AddFromFS("rsa-pss", jwt.SigningMethodPS256, fsys, "rsa-public.pem", "rsa-private.pem")
	j.AddFromFS("rsa", jwt.SigningMethodRS256, fsys, "rsa-public.pem", "rsa-private.pem")
	verifierMiddleware(a, s, j)
}

func readFile(a *assert.Assertion, fsys fs.FS, public, private string) ([]byte, []byte) {
	pub, err := fs.ReadFile(fsys, public)
	a.NotError(err).NotEmpty(pub)

	pvt, err := fs.ReadFile(fsys, private)
	a.NotError(err).NotEmpty(pvt)

	return pub, pvt
}

func verifierMiddleware(a *assert.Assertion, s web.Server, j *JWT[*testClaims]) {
	a.TB().Helper()

	const id = 1

	r := s.Routers().New("def", nil)
	r.Post("/login", func(ctx *web.Context) web.Responser {
		claims := &testClaims{
			ID:      id,
			Created: time.Now(),
		}
		return j.Render(ctx, http.StatusCreated, claims)
	})

	r.Post("/refresh", j.VerifiyRefresh(func(ctx *web.Context) web.Responser {
		a.TB().Helper()

		claims, ok := j.GetInfo(ctx)
		if !ok {
			return ctx.Problem(web.ProblemUnauthorized)
		}

		return j.Render(ctx, http.StatusCreated, &testClaims{ID: claims.ID, Created: ctx.Begin()})
	}))

	r.Get("/info", j.Middleware(func(ctx *web.Context) web.Responser {
		a.TB().Helper()

		val, found := j.GetInfo(ctx)
		if !found {
			return web.Status(http.StatusNotFound)
		}

		if val.ID != id {
			return web.Status(http.StatusUnauthorized)
		}

		return web.OK(nil)
	}))

	r.Delete("/login", j.Middleware(func(ctx *web.Context) web.Responser {
		a.TB().Helper()
		if err := j.Logout(ctx); err != nil {
			return ctx.Error(err, web.ProblemInternalServerError)
		}
		return web.NoContent()
	}))

	defer servertest.Run(a, s)()
	defer s.Close(0)

	servertest.Post(a, "http://localhost:8080/login", nil).
		Do(nil).
		Status(http.StatusCreated).BodyFunc(func(a *assert.Assertion, body []byte) {
		//a.TB().Helper()

		resp := &Response{}
		a.NotError(xjson.Unmarshal(bytes.NewBuffer(body), resp))
		a.NotEmpty(resp).
			NotEmpty(resp.Access).
			NotEmpty(resp.Refresh)

		servertest.Get(a, "http://localhost:8080/info").
			Header(mauth.AuthorizationHeader, prefix+resp.Access).
			Do(nil).
			Status(http.StatusOK)

		resp2 := &Response{}
		servertest.Post(a, "http://localhost:8080/refresh", nil).
			Header(mauth.AuthorizationHeader, prefix+resp.Refresh).
			Do(nil).
			Status(http.StatusCreated).
			BodyFunc(func(a *assert.Assertion, body []byte) {
				a.NotError(xjson.Unmarshal(bytes.NewBuffer(body), resp2)).
					NotEmpty(resp2).
					NotEmpty(resp2.Access).
					NotEmpty(resp2.Refresh)
			})

		a.True(j.v.blocker.TokenIsBlocked(resp.Access)).
			True(j.v.blocker.TokenIsBlocked(resp.Refresh)).
			False(j.v.blocker.TokenIsBlocked(resp2.Access)).
			False(j.v.blocker.TokenIsBlocked(resp2.Refresh))

		// 旧令牌已经无法访问
		servertest.Get(a, "http://localhost:8080/info").
			Header(mauth.AuthorizationHeader, prefix+resp.Access).
			Do(nil).
			Status(http.StatusUnauthorized)

		// 新令牌可以访问
		servertest.Get(a, "http://localhost:8080/info").
			Header(mauth.AuthorizationHeader, prefix+resp2.Access).
			Do(nil).
			Status(http.StatusOK)

		servertest.Delete(a, "http://localhost:8080/login").
			Header(mauth.AuthorizationHeader, prefix+resp2.Access).
			Do(nil).
			Status(http.StatusNoContent)

		// token 已经在 delete /login 中被弃用
		servertest.Get(a, "http://localhost:8080/info").
			Header(mauth.AuthorizationHeader, prefix+resp2.Access).
			Do(nil).
			Status(http.StatusUnauthorized)
	})
}

func TestVerifier_client(t *testing.T) {
	a := assert.New(t, false)
	s, j := newJWT(a, time.Hour, 2*time.Hour)
	j.AddRSAFromFS("rsa", jwt.SigningMethodRS256, os.DirFS("./testdata"), "rsa-public.pem", "rsa-private.pem")

	claims := &testClaims{
		ID: 1,
	}

	r := s.Routers().New("def", nil)
	r.Post("/login", func(ctx *web.Context) web.Responser {
		return j.Render(ctx, http.StatusCreated, claims)
	})

	r.Get("/info", j.Middleware(func(ctx *web.Context) web.Responser {
		val, found := j.GetInfo(ctx)
		if !found {
			return web.Status(http.StatusNotFound)
		}

		if val.ID != claims.ID {
			return web.Status(http.StatusUnauthorized)
		}

		return web.OK(nil)
	}))

	defer servertest.Run(a, s)()
	defer s.Close(0)

	servertest.Post(a, "http://localhost:8080/login", nil).
		Do(nil).
		Status(http.StatusCreated).BodyFunc(func(a *assert.Assertion, body []byte) {
		m := &Response{}
		a.NotError(json.Unmarshal(body, &m))
		a.NotEmpty(m).
			NotEmpty(m.Access).
			NotEmpty(m.Refresh)

		token, parts, err := jwt.NewParser().ParseUnverified(m.Access, &jwt.RegisteredClaims{})
		a.NotError(err).Equal(3, len(parts)).NotNil(token)
		header := decodeHeader(a, parts[0])
		a.Equal(header["alg"], "none").NotEmpty(header["kid"])

		// 改变 alg，影响
		header["alg"] = "ES256"
		parts[0] = encodeHeader(a, header)
		servertest.Get(a, "http://localhost:8080/info").
			Header(mauth.AuthorizationHeader, "BEARER "+strings.Join(parts, ".")).
			Do(nil).
			Status(http.StatusUnauthorized)

		// 改变 kid(kid 存在)，影响
		j.AddECDSAFromFS("ecdsa", jwt.SigningMethodES256, os.DirFS("./testdata"), "ec256-public.pem", "ec256-private.pem")
		header["kid"] = "ecdsa"
		header["alg"] = "none"
		parts[0] = encodeHeader(a, header)
		servertest.Get(a, "http://localhost:8080/info").
			Header(mauth.AuthorizationHeader, "BEARER "+strings.Join(parts, ".")).
			Do(nil).
			Status(http.StatusUnauthorized)

		// 改变 kid(kid 不存在)，影响
		header["kid"] = "not-exists"
		header["alg"] = "none"
		parts[0] = encodeHeader(a, header)
		servertest.Get(a, "http://localhost:8080/info").
			Header(mauth.AuthorizationHeader, "BEARER "+strings.Join(parts, ".")).
			Do(nil).
			Status(http.StatusUnauthorized)
	})

	s.Close(0)
}

func encodeHeader(a *assert.Assertion, header map[string]any) string {
	a.TB().Helper()

	data, err := json.Marshal(header)
	a.NotError(err).NotNil(data)

	return base64.RawURLEncoding.EncodeToString(data)
}

func decodeHeader(a *assert.Assertion, seg string) map[string]any {
	a.TB().Helper()

	data, err := base64.RawURLEncoding.DecodeString(seg)
	a.NotError(err).NotNil(data)

	header := make(map[string]any, 3)
	err = json.Unmarshal(data, &header)
	a.NotError(err)

	return header
}
