// SPDX-FileCopyrightText: 2024-2026 caixw
//
// SPDX-License-Identifier: MIT

package validator

import (
	"testing"

	"github.com/issue9/assert/v5"
)

func TestTimezone(t *testing.T) {
	a := assert.New(t, false)
	a.True(Timezone("Local")).
		True(Timezone("America/New_York")).
		False(Timezone("zzzz"))
}
