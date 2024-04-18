// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package locales

import (
	"os"
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/web"
)

var _ web.Plugin = &Locales{}

func TestLocales(t *testing.T) {
	a := assert.New(t, false)

	l := New("*.yaml")
	a.NotNil(l)

	l.Append(os.DirFS("./"))
	a.Length(l.f, 1)
}
