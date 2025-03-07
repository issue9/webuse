// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package validator

import (
	"testing"

	"github.com/issue9/assert/v4"
)

func TestSemver(t *testing.T) {
	a := assert.New(t, false)
	a.True(Semver("1.0.0")).
		False(Not(Semver)("1.0.0"))
}

func TestSemverCompatible(t *testing.T) {
	a := assert.New(t, false)

	v := SemverCompatible("1.0.0")
	a.True(v("1.1.1")).False(v("2.0.1"))
}

func TestSemverGreat(t *testing.T) {
	a := assert.New(t, false)

	v := SemverGreat("1.0.0")
	a.True(v("1.1.1")).False(v("1.0.0"))

	v = SemverGreatEqual("1.0.0")
	a.True(v("1.1.1")).True(v("1.0.0"))
}

func TestSemverLess(t *testing.T) {
	a := assert.New(t, false)

	v := SemverLess("2.0.0")
	a.True(v("1.1.1")).False(v("2.0.0"))

	v = SemverLessEqual("2.0.0")
	a.True(v("1.1.1")).True(v("2.0.0"))
}
