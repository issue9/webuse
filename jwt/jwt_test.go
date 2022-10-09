// SPDX-License-Identifier: MIT

package jwt

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/issue9/assert/v3"
	"github.com/issue9/web"
	"github.com/issue9/web/server/servertest"
)

type testClaims struct {
	ID      int64 `json:"id"`
	expires int64
	token   string
}

func (c *testClaims) SetExpired(t time.Duration) {
	c.expires = time.Now().Add(t).Unix()
}

func (c *testClaims) BuildRefresh(token string) Claims { return &testClaims{token: token} }

func (c *testClaims) Valid() error { return nil }

func newJWT(a *assert.Assertion, expired, refresh time.Duration) (*JWT[*testClaims], *memoryBlocker) {
	m := &memoryBlocker{}
	b := func() *testClaims {
		return &testClaims{}
	}
	j := New[*testClaims](m, b, expired, refresh, nil)
	a.NotNil(j)
	return j, m
}

func TestVerifier_Middleware(t *testing.T) {
	a := assert.New(t, false)

	j, m := newJWT(a, time.Hour, 0)
	j.AddHMAC("hmac-secret", jwt.SigningMethodHS256, []byte("secret"))
	verifierMiddleware(a, j, m)

	a.PanicString(func() {
		j.AddHMAC("hmac-secret", jwt.SigningMethodHS256, []byte("secret"))
	}, "存在同名的签名方法 hmac-secret")

	a.PanicString(func() {
		j.s.AddHMAC("hmac-secret", jwt.SigningMethodHS256, []byte("secret"))
	}, "存在同名的签名方法 hmac-secret")

	j, m = newJWT(a, time.Hour, 0)
	j.AddRSAFromFS("rsa", jwt.SigningMethodRS256, os.DirFS("./testdata"), "rsa-public.pem", "rsa-private.pem")
	verifierMiddleware(a, j, m)

	j, m = newJWT(a, time.Hour, 0)
	j.AddRSAPSSFromFS("rsa-pss", jwt.SigningMethodPS256, os.DirFS("./testdata"), "rsa-public.pem", "rsa-private.pem")
	verifierMiddleware(a, j, m)

	j, m = newJWT(a, time.Hour, 0)
	j.AddECDSAFromFS("ecdsa", jwt.SigningMethodES256, os.DirFS("./testdata"), "ec256-public.pem", "ec256-private.pem")
	verifierMiddleware(a, j, m)

	j, m = newJWT(a, time.Hour, 0)
	j.AddEd25519FromFS("ed25519", jwt.SigningMethodEdDSA, os.DirFS("./testdata"), "ed25519-public.pem", "ed25519-private.pem")
	verifierMiddleware(a, j, m)

	j, m = newJWT(a, time.Hour, 0)
	j.AddEd25519FromFS("ed25519", jwt.SigningMethodEdDSA, os.DirFS("./testdata"), "ed25519-public.pem", "ed25519-private.pem")
	j.AddECDSAFromFS("ecdsa", jwt.SigningMethodES256, os.DirFS("./testdata"), "ec256-public.pem", "ec256-private.pem")
	j.AddRSAPSSFromFS("rsa-pss", jwt.SigningMethodPS256, os.DirFS("./testdata"), "rsa-public.pem", "rsa-private.pem")
	j.AddRSAFromFS("rsa", jwt.SigningMethodRS256, os.DirFS("./testdata"), "rsa-public.pem", "rsa-private.pem")
	verifierMiddleware(a, j, m)
	verifierMiddleware(a, j, m)
	verifierMiddleware(a, j, m)
	verifierMiddleware(a, j, m)
}

func verifierMiddleware(a *assert.Assertion, j *JWT[*testClaims], d *memoryBlocker) {
	a.TB().Helper()
	d.clear()

	claims := &testClaims{
		ID: 1,
	}

	s := servertest.NewTester(a, nil)
	r := s.Router()
	r.Post("/login", func(ctx *web.Context) web.Responser {
		return j.Render(ctx, http.StatusCreated, claims)
	})

	r.Get("/info", j.Middleware(func(ctx *web.Context) web.Responser {
		val, found := j.GetValue(ctx)
		if !found {
			return web.Status(http.StatusNotFound)
		}

		if val.ID != claims.ID {
			return web.Status(http.StatusUnauthorized)
		}

		return web.OK(nil)
	}))

	r.Delete("/login", j.Middleware(func(ctx *web.Context) web.Responser {
		val, found := j.GetValue(ctx)
		if !found {
			return web.Status(http.StatusNotFound)
		}

		if val.ID != claims.ID {
			return web.Status(http.StatusUnauthorized)
		}

		if d != nil {
			d.BlockToken(j.GetToken(ctx))
		}

		return web.NoContent()
	}))

	s.GoServe()

	s.NewRequest(http.MethodPost, "/login", nil).
		Do(nil).
		Status(http.StatusCreated).BodyFunc(func(a *assert.Assertion, body []byte) {
		resp := &Response{}
		a.NotError(json.Unmarshal(body, resp))
		a.NotEmpty(resp).
			NotEmpty(resp.Access).
			Zero(resp.Refresh)

		s.Get("/info").
			Header("Authorization", "BEARER "+resp.Access).
			Do(nil).
			Status(http.StatusOK)

		s.Delete("/login").
			Header("Authorization", resp.Access).
			Do(nil).
			Status(http.StatusNoContent)

		// token 已经在 delete /login 中被弃用
		s.Get("/info").
			Header("Authorization", resp.Access).
			Do(nil).
			Status(http.StatusUnauthorized)
	})

	s.Close(0)
}

func TestVerifier_client(t *testing.T) {
	a := assert.New(t, false)
	j, _ := newJWT(a, time.Hour, 2*time.Hour)
	j.AddRSAFromFS("rsa", jwt.SigningMethodRS256, os.DirFS("./testdata"), "rsa-public.pem", "rsa-private.pem")

	claims := &testClaims{
		ID: 1,
	}

	s := servertest.NewTester(a, nil)
	r := s.Router()
	r.Post("/login", func(ctx *web.Context) web.Responser {
		return j.Render(ctx, http.StatusCreated, claims)
	})

	r.Get("/info", j.Middleware(func(ctx *web.Context) web.Responser {
		val, found := j.GetValue(ctx)
		if !found {
			return web.Status(http.StatusNotFound)
		}

		if val.ID != claims.ID {
			return web.Status(http.StatusUnauthorized)
		}

		return web.OK(nil)
	}))

	s.GoServe()

	s.NewRequest(http.MethodPost, "/login", nil).
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
		s.Get("/info").
			Header("Authorization", "BEARER "+strings.Join(parts, ".")).
			Do(nil).
			Status(http.StatusInternalServerError)

		// 改变 kid(kid 存在)，影响
		j.AddECDSAFromFS("ecdsa", jwt.SigningMethodES256, os.DirFS("./testdata"), "ec256-public.pem", "ec256-private.pem")
		header["kid"] = "ecdsa"
		header["alg"] = "none"
		parts[0] = encodeHeader(a, header)
		s.Get("/info").
			Header("Authorization", "BEARER "+strings.Join(parts, ".")).
			Do(nil).
			Status(http.StatusInternalServerError)

		// 改变 kid(kid 不存在)，影响
		header["kid"] = "not-exists"
		header["alg"] = "none"
		parts[0] = encodeHeader(a, header)
		s.Get("/info").
			Header("Authorization", "BEARER "+strings.Join(parts, ".")).
			Do(nil).
			Status(http.StatusInternalServerError)
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

	var data []byte
	var err error
	if jwt.DecodePaddingAllowed {
		if l := len(seg) % 4; l > 0 {
			seg += strings.Repeat("=", 4-l)
		}
		data, err = base64.URLEncoding.DecodeString(seg)
	} else {
		data, err = base64.RawURLEncoding.DecodeString(seg)
	}
	a.NotError(err).NotNil(data)

	header := make(map[string]any, 3)
	err = json.Unmarshal(data, &header)
	a.NotError(err)

	return header
}
