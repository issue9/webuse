// SPDX-License-Identifier: MIT

package jwt

import "github.com/golang-jwt/jwt/v4"

func claimsBuilder() jwt.Claims { return &jwt.RegisteredClaims{} }
