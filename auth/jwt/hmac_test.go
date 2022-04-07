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

var _ web.Middleware = &hmac{}

func TestHMAC_Sign(t *testing.T) {
	a := assert.New(t, false)
	j := New(claimsBuilder, 5*time.Minute)
	a.NotNil(j)
	m := j.NewHMAC([]byte("abc"), jwt.SigningMethodHS256)
	a.NotNil(m)

	token, err := m.Sign(jwt.RegisteredClaims{
		Issuer:  "issuer",
		Subject: "subject",
		ID:      "id",
	})
	a.NotError(err).NotEmpty(token)
}

func TestHMAC_Middleware(t *testing.T) {
	a := assert.New(t, false)
	j := New(claimsBuilder, 5*time.Minute)
	a.NotNil(j)
	m := j.NewHMAC([]byte("secret"), jwt.SigningMethodHS256)
	a.NotNil(m)

	claims := jwt.RegisteredClaims{
		Issuer:  "issuer",
		Subject: "subject",
		ID:      "id",
	}

	s := servertest.NewTester(a, nil)
	r := s.NewRouter()
	r.Post("/login", func(ctx *web.Context) web.Responser {
		token, err := m.Sign(claims)
		if err != nil {
			return ctx.InternalServerError(err)
		}
		return ctx.Created(map[string]string{"token": token}, "")
	})

	r.Delete("/login", m.Middleware(func(ctx *web.Context) web.Responser {
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
