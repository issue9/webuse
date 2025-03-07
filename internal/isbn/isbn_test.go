// SPDX-FileCopyrightText: 2022-2024 caixw
//
// SPDX-License-Identifier: MIT

package isbn

import (
	"testing"

	"github.com/issue9/assert/v4"
)

func TestEraseMinus(t *testing.T) {
	a := assert.New(t, false)

	a.Equal(eraseMinus([]byte("abc-def-abc-")), []byte("abcdefabc"))
	a.Equal(eraseMinus([]byte("abc-def-abc")), []byte("abcdefabc"))
	a.Equal(eraseMinus([]byte("-_abc-def-abc-")), []byte("_abcdefabc"))
	a.Equal(eraseMinus([]byte("-abc-d_ef-abc-")), []byte("abcd_efabc"))
}

func TestISBN(t *testing.T) {
	a := assert.New(t, false)

	a.True(IsValid([]byte("1-919876-03-0")))
	a.True(IsValid([]byte("0-471-00084-1")))
	a.True(IsValid([]byte("978-7-301-04815-3")))
	a.True(IsValid([]byte("978-986-181-728-6")))
}
