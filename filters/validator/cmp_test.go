// SPDX-FileCopyrightText: 2022-2024 caixw
//
// SPDX-License-Identifier: MIT

package validator

import (
	"math"
	"testing"

	"github.com/issue9/assert/v4"
)

func TestRange(t *testing.T) {
	a := assert.New(t, false)

	a.Panic(func() {
		Between(100, 5)
	})

	a.True(BetweenEqual(5, math.MaxInt16)(5))
	a.True(BetweenEqual(5.0, 100.0)(5.1))
	a.False(BetweenEqual(5, 100)(200))
	a.False(BetweenEqual(5, 100)(-1))
	a.False(BetweenEqual(5.0, 100.0)(-1.1))

	r := GreatEqual(6)
	a.True(r(6))
	a.True(r(10))
	a.False(r(5))

	r = LessEqual(6)
	a.True(r(6))
	a.False(r(10))
	a.True(r(5))
}
