// SPDX-License-Identifier: MIT

package jwt

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v4"
	"github.com/issue9/assert/v2"
	"github.com/issue9/web"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/servertest"
)

func TestJWT_Middleware(t *testing.T) {
	a := assert.New(t, false)

	j, m := newJWT(a)
	j.AddHMAC("hmac-secret", jwt.SigningMethodHS256, []byte("secret"))
	testJWT_Middleware(a, j, m)

	a.PanicString(func() {
		j.AddHMAC("hmac-secret", jwt.SigningMethodHS256, []byte("secret"))
	}, "存在同名的签名方法 hmac-secret")

	j, m = newJWT(a)
	j.AddRSAFromFS("rsa", jwt.SigningMethodRS256, os.DirFS("./testdata"), "rsa-private.pem", "rsa-public.pem")
	testJWT_Middleware(a, j, m)

	j, m = newJWT(a)
	j.AddRSAPSSFromFS("rsa-pss", jwt.SigningMethodPS256, os.DirFS("./testdata"), "rsa-private.pem", "rsa-public.pem")
	testJWT_Middleware(a, j, m)

	j, m = newJWT(a)
	j.AddECDSAFromFS("ecdsa", jwt.SigningMethodES256, os.DirFS("./testdata"), "ec256-private.pem", "ec256-public.pem")
	testJWT_Middleware(a, j, m)

	j, m = newJWT(a)
	j.AddEd25519FromFS("ed25519", jwt.SigningMethodEdDSA, os.DirFS("./testdata"), "ed25519-private.pem", "ed25519-public.pem")
	testJWT_Middleware(a, j, m)
}

func TestJWT_keyIDs(t *testing.T) {
	a := assert.New(t, false)

	j, m := newJWT(a)
	j.AddHMAC("hmac-secret", jwt.SigningMethodHS256, []byte("secret"))
	j.AddRSAFromFS("rsa", jwt.SigningMethodRS256, os.DirFS("./testdata"), "rsa-private.pem", "rsa-public.pem")
	j.AddECDSAFromFS("ecdsa", jwt.SigningMethodES256, os.DirFS("./testdata"), "ec256-private.pem", "ec256-public.pem")

	a.Equal(j.KeyIDs(), []string{"hmac-secret", "rsa", "ecdsa"})

	testJWT_Middleware(a, j, m)
	testJWT_Middleware(a, j, m)
	testJWT_Middleware(a, j, m)
}

func newJWT(a *assert.Assertion) (*stdJWT, *memoryDiscarder) {
	m := &memoryDiscarder{}
	j := New[*jwt.RegisteredClaims](m, claimsBuilder)
	a.NotNil(j)
	return j, m
}

func testJWT_Middleware(a *assert.Assertion, j *stdJWT, d *memoryDiscarder) {
	claims := jwt.RegisteredClaims{
		Issuer:  "issuer",
		Subject: "subject",
		ID:      "id",
	}

	s := servertest.NewTester(a, nil)
	r := s.NewRouter()
	r.Post("/login", func(ctx *web.Context) web.Responser {
		token, err := j.Sign(claims)
		if err != nil {
			return ctx.InternalServerError(err)
		}
		return ctx.Created(map[string]string{"token": token}, "")
	})

	r.Get("/info", j.Middleware(func(ctx *server.Context) server.Responser {
		val, found := j.GetValue(ctx)
		if !found {
			return ctx.Status(http.StatusNotFound)
		}

		if val.Issuer != claims.Issuer || val.Subject != claims.Subject || val.ID != claims.ID {
			return ctx.Status(http.StatusUnauthorized)
		}

		return ctx.OK(nil)
	}))

	r.Delete("/login", j.Middleware(func(ctx *web.Context) web.Responser {
		val, found := j.GetValue(ctx)
		if !found {
			return ctx.Status(http.StatusNotFound)
		}

		if val.Issuer != claims.Issuer || val.Subject != claims.Subject || val.ID != claims.ID {
			return ctx.Status(http.StatusUnauthorized)
		}

		if d != nil {
			d.DiscardToken(j.GetToken(ctx))
		}

		return ctx.NoContent()
	}))

	s.GoServe()

	s.NewRequest(http.MethodPost, "/login", nil).
		Do(nil).
		Status(http.StatusCreated).BodyFunc(func(a *assert.Assertion, body []byte) {
		m := map[string]string{}
		a.NotError(json.Unmarshal(body, &m))
		a.NotEmpty(m).
			NotEmpty(m["token"])

		s.Get("/info").
			Header("Authorization", "BEARER "+m["token"]).
			Do(nil).
			Status(http.StatusOK)

		s.Delete("/login").
			Header("Authorization", m["token"]).
			Do(nil).
			Status(http.StatusNoContent)

		// token 已经在 delete /login 中被弃用
		s.Get("/info").
			Header("Authorization", m["token"]).
			Do(nil).
			Status(http.StatusUnauthorized)
	})

	s.Close(0)
	s.Wait()
}

func TestJWT_client(t *testing.T) {
	a := assert.New(t, false)
	j, _ := newJWT(a)
	j.AddRSAFromFS("rsa", jwt.SigningMethodRS256, os.DirFS("./testdata"), "rsa-private.pem", "rsa-public.pem")

	claims := jwt.RegisteredClaims{
		Issuer:  "issuer",
		Subject: "subject",
		ID:      "id",
	}

	s := servertest.NewTester(a, nil)
	r := s.NewRouter()
	r.Post("/login", func(ctx *web.Context) web.Responser {
		token, err := j.Sign(claims)
		if err != nil {
			return ctx.InternalServerError(err)
		}
		return ctx.Created(map[string]string{"token": token}, "")
	})

	r.Get("/info", j.Middleware(func(ctx *server.Context) server.Responser {
		val, found := j.GetValue(ctx)
		if !found {
			return ctx.Status(http.StatusNotFound)
		}

		if val.Issuer != claims.Issuer || val.Subject != claims.Subject || val.ID != claims.ID {
			return ctx.Status(http.StatusUnauthorized)
		}

		return ctx.OK(nil)
	}))

	s.GoServe()

	s.NewRequest(http.MethodPost, "/login", nil).
		Do(nil).
		Status(http.StatusCreated).BodyFunc(func(a *assert.Assertion, body []byte) {
		m := map[string]string{}
		a.NotError(json.Unmarshal(body, &m))
		a.NotEmpty(m).
			NotEmpty(m["token"])

		token, parts, err := jwt.NewParser().ParseUnverified(m["token"], &jwt.RegisteredClaims{})
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
		j.AddECDSAFromFS("ecdsa", jwt.SigningMethodES256, os.DirFS("./testdata"), "ec256-private.pem", "ec256-public.pem")
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
	s.Wait()
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
