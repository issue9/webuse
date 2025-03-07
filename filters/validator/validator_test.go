// SPDX-FileCopyrightText: 2022-2024 caixw
//
// SPDX-License-Identifier: MIT

package validator

import (
	"testing"

	"github.com/issue9/assert/v4"
)

func TestNot(t *testing.T) {
	a := assert.New(t, false)

	a.True(Zero(0)).False(Zero(1))

	nz := Not[int](Zero[int])
	a.False(nz(0)).True(nz(1))
}

func TestAnd_Or(t *testing.T) {
	a := assert.New(t, false)

	and := And(Between(0, 100), Between(-1, 50))
	a.False(and(0)).
		True(and(1)).
		False(and(51))

	or := Or(Between(0, 100), Between(-1, 50))
	a.True(or(0)).
		True(or(1)).
		False(or(500))
}
