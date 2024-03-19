// SPDX-FileCopyrightText: 2015-2024 caixw
//
// SPDX-License-Identifier: MIT

package jwt

import (
	"github.com/issue9/web"

	"github.com/issue9/webuse/v7/middlewares/auth"
)

var (
	_ web.MiddlewareFunc = (&Verifier[*testClaims]{}).VerifyRefresh
	_ auth.Auth          = &Verifier[*testClaims]{}
)
