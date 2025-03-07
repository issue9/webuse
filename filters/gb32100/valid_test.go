// SPDX-FileCopyrightText: 2022-2024 caixw
//
// SPDX-License-Identifier: MIT

package gb32100

import (
	"testing"

	"github.com/issue9/assert/v4"
)

func TestCodes(t *testing.T) {
	a := assert.New(t, false)

	a.Equal(31, len(codeIndexes))
	a.Equal(17, len(factors))
}

func TestIsValid(t *testing.T) {
	a := assert.New(t, false)

	for i, item := range validData {
		a.True(IsValid([]byte(item)), "failed @ %d:%s", i, item)
	}
}
