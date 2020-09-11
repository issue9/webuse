// SPDX-License-Identifier: MIT

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
