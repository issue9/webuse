// SPDX-License-Identifier: MIT

package jwt

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/issue9/sliceutil"
)

var _ Discarder[*jwt.RegisteredClaims] = &memoryDiscarder{}

type memoryDiscarder struct {
	tokens []string
	claims []string
}

func (m *memoryDiscarder) IsDiscarded(t string) bool {
	return sliceutil.Exists(m.tokens, func(e string) bool { return e == t })
}

func (m *memoryDiscarder) ClaimsIsDiscarded(t *jwt.RegisteredClaims) bool {
	return sliceutil.Exists(m.claims, func(e string) bool { return e == t.ID })
}

func (m *memoryDiscarder) Discard(t string) { m.tokens = append(m.tokens, t) }

func (m *memoryDiscarder) DiscardClaims(t *jwt.RegisteredClaims) { m.claims = append(m.claims, t.ID) }
