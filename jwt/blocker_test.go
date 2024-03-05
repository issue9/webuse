// SPDX-FileCopyrightText: 2015-2024 caixw
//
// SPDX-License-Identifier: MIT

package jwt

import (
	"slices"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
)

var _ Blocker[*testClaims] = &memoryBlocker{}

type memoryBlocker struct {
	tokens []string
	claims []string
}

func (m *memoryBlocker) TokenIsBlocked(t string) bool {
	return slices.Index(m.tokens, t) >= 0
}

func (m *memoryBlocker) ClaimsIsBlocked(t *testClaims) bool {
	return slices.Index(m.claims, strconv.FormatInt(t.ID, 10)) >= 0
}

func (m *memoryBlocker) BlockToken(t string) { m.tokens = append(m.tokens, t) }

func (m *memoryBlocker) BlockClaims(t *jwt.RegisteredClaims) { m.claims = append(m.claims, t.ID) }

func (m *memoryBlocker) clear() {
	m.tokens = m.tokens[:0]
	m.claims = m.claims[:0]
}
