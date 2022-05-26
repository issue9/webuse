// SPDX-License-Identifier: MIT

package jwt

import (
	"testing"

	"github.com/golang-jwt/jwt/v4"
	"github.com/issue9/assert/v2"
	"github.com/issue9/web"
)

var _ web.Middleware = &stdJWT{}

type stdJWT = JWT[*jwt.RegisteredClaims]

func claimsBuilder() *jwt.RegisteredClaims { return &jwt.RegisteredClaims{} }

func TestNew(t *testing.T) {
	a := assert.New(t, false)
	j := New(nil, claimsBuilder)
	a.NotNil(j).
		NotNil(j.discarder).
		NotNil(j.claimsBuilder).
		NotNil(j.keyFunc)
}
