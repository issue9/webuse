// SPDX-License-Identifier: MIT

package jwt

import (
	"encoding/json"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/issue9/assert/v2"
	"github.com/issue9/web"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/servertest"
)

var _ web.Middleware = &stdJWT{}

type stdJWT = JWT[*jwt.RegisteredClaims]

func claimsBuilder() *jwt.RegisteredClaims { return &jwt.RegisteredClaims{} }

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
