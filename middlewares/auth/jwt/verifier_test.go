// SPDX-FileCopyrightText: 2015-2024 caixw
//
// SPDX-License-Identifier: MIT

package jwt

import "github.com/issue9/web"

var (
	_ web.Middleware     = &Verifier[*testClaims]{}
	_ web.MiddlewareFunc = (&Verifier[*testClaims]{}).VerifiyRefresh
)
