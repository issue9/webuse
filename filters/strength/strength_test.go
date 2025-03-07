// SPDX-FileCopyrightText: 2022-2025 caixw
//
// SPDX-License-Identifier: MIT

package strength

import (
	"testing"

	"github.com/issue9/assert/v4"
)

func TestStrength_Gen(t *testing.T) {
	a := assert.New(t, false)
	v := &Strength{Length: 5}
	a.True(v.Valid(string(v.Gen())))

	v = &Strength{Length: 5, Number: 5}
	a.True(v.Valid(string(v.Gen())))

	v = &Strength{Length: 5, Number: 5, Upper: 3}
	a.True(v.Valid(string(v.Gen())))

	v = &Strength{Length: 5, Number: 5, Lower: 3}
	a.True(v.Valid(string(v.Gen())))

	v = &Strength{Length: 5, Number: 5, Lower: 3, Punct: 3}
	a.True(v.Valid(string(v.Gen())))
}

func TestStrength_Valid(t *testing.T) {
	a := assert.New(t, false)

	v := &Strength{}
	a.True(v.Valid(""))
	a.True(v.Valid("123"))

	v = &Strength{Length: 3}
	a.False(v.Valid(""))
	a.False(v.Valid("12"))
	a.True(v.Valid("123"))
	a.True(v.Valid("Abcdef"))

	v = &Strength{Length: 3, Upper: 2}
	a.False(v.Valid(""))
	a.False(v.Valid("12"))
	a.False(v.Valid("123"))
	a.False(v.Valid("123A"))
	a.True(v.Valid("123AB"))
	a.False(v.Valid("AB"))
	a.True(v.Valid("ABc"))

	v = &Strength{Upper: 2, Lower: 3, Punct: 1}
	a.False(v.Valid(""))
	a.False(v.Valid("12345678"))
	a.True(v.Valid("ABcde>"))
	a.True(v.Valid("ABcde123><"))

	v = &Strength{Length: 4, Upper: 2, Punct: 1}
	a.False(v.Valid(""))
	a.False(v.Valid("12345678"))
	a.False(v.Valid("AB>"))
	a.True(v.Valid("AB=!>"))
}
