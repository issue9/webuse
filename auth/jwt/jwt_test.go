// SPDX-License-Identifier: MIT

package jwt

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/issue9/assert/v2"
	"github.com/issue9/web"
	"github.com/issue9/web/server/servertest"

	"github.com/issue9/middleware/v6/auth"
)

var _ web.Middleware = &JWT{}

func claimsBuilder() jwt.Claims { return &jwt.RegisteredClaims{} }

func TestJWT_Sign(t *testing.T) {
	a := assert.New(t, false)
	j := New(claimsBuilder, "abc", 5*time.Minute, jwt.SigningMethodHS256)
	a.NotNil(j)

	token, err := j.Sign(jwt.RegisteredClaims{
		Issuer:  "issuer",
		Subject: "subject",
		ID:      "id",
	})
	a.NotError(err).NotEmpty(token)
}

func TestJWT_Middleware(t *testing.T) {
	a := assert.New(t, false)
	j := New(claimsBuilder, "secret", 5*time.Minute, jwt.SigningMethodHS256)
	a.NotNil(j)
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

	r.Delete("/login", j.Middleware(func(ctx *web.Context) web.Responser {
		val, found := auth.GetValue(ctx)
		if !found {
			return ctx.Status(http.StatusNotFound)
		}
		v, ok := val.(*jwt.RegisteredClaims)
		if !ok {
			return ctx.Status(http.StatusInternalServerError)
		}

		if v.Issuer != claims.Issuer || v.Subject != claims.Subject || v.ID != claims.ID {
			return ctx.Status(http.StatusUnauthorized)
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

		s.Delete("/login").
			Header("Authorization", m["token"]).
			Do(nil).
			Status(http.StatusNoContent)
	})

	s.Close(0)
	s.Wait()

}
