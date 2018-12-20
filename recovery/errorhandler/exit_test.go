// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package errorhandler

import (
	"net/http"
	"testing"

	"github.com/issue9/assert"
)

func TestExit(t *testing.T) {
	a := assert.New(t)

	a.Panic(func() { Exit(5) })
	a.Panic(func() { Exit(0) })

	func() {
		defer func() {
			msg := recover()
			val, ok := msg.(httpStatus)
			a.True(ok).Equal(val, http.StatusNotFound)
		}()

		Exit(http.StatusNotFound)
	}()
}
