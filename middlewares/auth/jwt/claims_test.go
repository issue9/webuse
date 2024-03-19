// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package jwt

import "github.com/golang-jwt/jwt/v5"

type testClaims struct {
	jwt.MapClaims
	ID    int64 `json:"id"`
	token string
}

func (c *testClaims) BaseToken() string { return c.token }

func (c *testClaims) BuildRefresh(token string) Claims { return &testClaims{token: token} }

func (c *testClaims) Valid() error { return nil }
