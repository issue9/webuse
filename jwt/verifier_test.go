// SPDX-License-Identifier: MIT

package jwt

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/issue9/web"
)

var _ web.Middleware = &stdVerifier{}

type stdVerifier = Verifier[*jwt.RegisteredClaims]
