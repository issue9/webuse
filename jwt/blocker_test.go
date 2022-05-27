// SPDX-License-Identifier: MIT

package jwt

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/issue9/sliceutil"
)

var _ Blocker[*jwt.RegisteredClaims] = &memoryBlocker{}

type memoryBlocker struct {
	tokens []string
	claims []string
}

func (m *memoryBlocker) TokenIsBlocked(t string) bool {
	return sliceutil.Exists(m.tokens, func(e string) bool { return e == t })
}

func (m *memoryBlocker) ClaimsIsBlocked(t *jwt.RegisteredClaims) bool {
	return sliceutil.Exists(m.claims, func(e string) bool { return e == t.ID })
}

func (m *memoryBlocker) BlockToken(t string) { m.tokens = append(m.tokens, t) }

func (m *memoryBlocker) BlockClaims(t *jwt.RegisteredClaims) { m.claims = append(m.claims, t.ID) }

func (m *memoryBlocker) clear() {
	m.tokens = m.tokens[:0]
	m.claims = m.claims[:0]
}
