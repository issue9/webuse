// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package sanitizer

import (
	"testing"

	"github.com/issue9/assert/v4"
)

func TestAbs(t *testing.T) {
	a := assert.New(t, false)

	v1 := -1.0
	Abs(&v1)
	a.Equal(v1, 1)

	v1 = 1.0
	Abs(&v1)
	a.Equal(v1, 1)

	v1 = 1.555
	Abs(&v1)
	a.Equal(v1, 1.555)

	v1 = 0
	Abs(&v1)
	a.Equal(v1, 0)
}
