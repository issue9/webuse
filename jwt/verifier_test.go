// SPDX-License-Identifier: MIT

package jwt

import "github.com/issue9/web"

var _ web.Middleware = &stdVerifier{}

type stdVerifier = Verifier[*testClaims]
