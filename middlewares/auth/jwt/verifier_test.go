// SPDX-FileCopyrightText: 2015-2024 caixw
//
// SPDX-License-Identifier: MIT

package jwt

import "github.com/issue9/webuse/v7/middlewares/auth"

var _ auth.Auth[*testClaims] = &Verifier[*testClaims]{}
