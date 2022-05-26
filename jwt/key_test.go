// SPDX-License-Identifier: MIT

package jwt

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/golang-jwt/jwt/v4"
	"github.com/issue9/assert/v2"
	"github.com/issue9/web"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/servertest"
)

func TestJWT_AddHMAC(t *testing.T) {
	a := assert.New(t, false)
	j, m := newJWT(a)

	j.AddHMAC("hmac-secret", jwt.SigningMethodHS256, []byte("secret"))
	testJWT_Middleware(a, j, m, "hmac-secret")

	a.PanicString(func() {
		j.AddHMAC("hmac-secret", jwt.SigningMethodHS256, []byte("secret"))
	}, "存在同名的签名方法 hmac-secret")
}

func TestJWT_AddFS(t *testing.T) {
	a := assert.New(t, false)
	j, m := newJWT(a)

	j.AddRSAFromFS("rsa", jwt.SigningMethodRS256, os.DirFS("./testdata"), "rsa-private.pem", "rsa-public.pem")
	testJWT_Middleware(a, j, m, "rsa")

	j.AddRSAPSSFromFS("rsa-pss", jwt.SigningMethodPS256, os.DirFS("./testdata"), "rsa-private.pem", "rsa-public.pem")
	j.AddECDSAFromFS("ecdsa", jwt.SigningMethodES256, os.DirFS("./testdata"), "ec256-private.pem", "ec256-public.pem")
	j.AddEd25519FromFS("ed25519", jwt.SigningMethodEdDSA, os.DirFS("./testdata"), "ed25519-private.pem", "ed25519-public.pem")

	testJWT_Middleware(a, j, m, "rsa-pss")
	testJWT_Middleware(a, j, m, "ecdsa")
	testJWT_Middleware(a, j, m, "ed25519")

	a.Equal(j.KeyIDs(), []string{"rsa", "rsa-pss", "ecdsa", "ed25519"})
}

func newJWT(a *assert.Assertion) (*stdJWT, *memoryDiscarder) {
	m := &memoryDiscarder{}
	j := New[*jwt.RegisteredClaims](m, claimsBuilder)
	a.NotNil(j)
	return j, m
}

func testJWT_Middleware(a *assert.Assertion, j *stdJWT, d *memoryDiscarder, kid string) {
	claims := jwt.RegisteredClaims{
		Issuer:  "issuer",
		Subject: "subject",
		ID:      "id",
	}

	s := servertest.NewTester(a, nil)
	r := s.NewRouter()
	r.Post("/login", func(ctx *web.Context) web.Responser {
		token, err := j.Sign(claims, map[string]any{"kid": kid})
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
