// SPDX-License-Identifier: MIT

package jwt

import "github.com/golang-jwt/jwt/v4"

var _ Responser = &response{}

type stdSigner = Signer[*jwt.RegisteredClaims]
