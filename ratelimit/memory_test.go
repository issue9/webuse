// SPDX-License-Identifier: MIT

package ratelimit

import (
	"testing"

	"github.com/issue9/assert"
)

var _ Store = &memory{}

func TestMemory(t *testing.T) {
	a := assert.New(t)
	s := NewMemory(10)
	a.NotNil(s)

	b1 := newBucket(10, 20)
	b2 := newBucket(10, 20)

	a.Nil(s.Get("b1"))

	s.Set("b1", b1)
	a.Equal(b1, s.Get("b1"))

	s.Set("b1", b2)
	a.Equal(b2, s.Get("b1"))

	a.NotError(s.Delete("b1"))
	a.NotError(s.Delete("b1")) // 空
}
