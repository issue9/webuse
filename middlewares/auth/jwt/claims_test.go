// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/issue9/web"
)

type testClaims struct {
	jwt.MapClaims
	ID      int64     `json:"id"`
	Created time.Time `json:"created"`
	Token   string    //`json:"token"`
}

func (c *testClaims) BaseToken() string { return c.Token }

func (c *testClaims) BuildRefresh(token string, ctx *web.Context) Claims {
	return &testClaims{Token: token, Created: ctx.Begin(), ID: c.ID}
}

func (c *testClaims) Valid() error { return nil }
