// SPDX-License-Identifier: MIT

package jwt

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/golang-jwt/jwt/v4"
	"github.com/issue9/assert/v2"
	"github.com/issue9/web"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/servertest"
)

var _ web.Middleware = &JWT[*jwt.RegisteredClaims]{}

func claimsBuilder() *jwt.RegisteredClaims { return &jwt.RegisteredClaims{} }

func getDefaultClaims() jwt.RegisteredClaims {
	return jwt.RegisteredClaims{
		Issuer:  "issuer",
		Subject: "subject",
		ID:      "id",
	}
}

func TestHMAC(t *testing.T) {
	a := assert.New(t, false)

	j := NewHMAC[*jwt.RegisteredClaims](&memoryDiscarder{}, claimsBuilder, jwt.SigningMethodHS256, []byte("abc"))
	a.NotNil(j)
	testJWT_Sign(a, j)

	m := &memoryDiscarder{}
	j = NewHMAC[*jwt.RegisteredClaims](m, claimsBuilder, jwt.SigningMethodHS256, []byte("secret"))
	a.NotNil(j)
	testJWT_Middleware(a, j, m)
}

func testJWT_Sign(a *assert.Assertion, j *JWT[*jwt.RegisteredClaims]) {
	token, err := j.Sign(getDefaultClaims())
	a.NotError(err).NotEmpty(token)
}

func testJWT_Middleware(a *assert.Assertion, j *JWT[*jwt.RegisteredClaims], d *memoryDiscarder) {
	claims := getDefaultClaims()

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
			d.Discard(j.GetToken(ctx))
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
