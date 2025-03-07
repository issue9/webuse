// SPDX-FileCopyrightText: 2022-2024 caixw
//
// SPDX-License-Identifier: MIT

package luhn

import (
	"testing"

	"github.com/issue9/assert/v4"
)

func TestIsValid(t *testing.T) {
	a := assert.New(t, false)

	a.True(IsValid([]byte("6259650871772098")))
	a.True(IsValid([]byte("79927398713")))
	a.False(IsValid([]byte("79927398710")))
}

func TestBuild(t *testing.T) {
	a := assert.New(t, false)

	a.Equal("6259650871772098", string(Build([]byte("625965087177209"))))
	a.True(IsValid(Build([]byte("625965087177209"))))
}
