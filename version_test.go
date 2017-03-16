// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"testing"

	"github.com/issue9/assert"
)

func TestFindVersionNumber(t *testing.T) {
	a := assert.New(t)

	a.Equal(findVersionNumber(""), "")
	a.Equal(findVersionNumber("version="), "")
	a.Equal(findVersionNumber("Version="), "")
	a.Equal(findVersionNumber(";version="), "")
	a.Equal(findVersionNumber(";version=;"), "")
	a.Equal(findVersionNumber(";version=1.0"), "1.0")
	a.Equal(findVersionNumber(";version=1.0;"), "1.0")
	a.Equal(findVersionNumber(";version=1.0;application/json"), "1.0")
	a.Equal(findVersionNumber("application/json;version=1.0"), "1.0")
	a.Equal(findVersionNumber("application/json;version=1.0;application/json"), "1.0")
}
