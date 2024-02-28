// SPDX-License-Identifier: MIT

package jwt

import (
	"testing"
	"time"

	"github.com/issue9/assert/v4"
)

func TestNewSigner(t *testing.T) {
	a := assert.New(t, false)

	a.PanicString(func() {
		NewSigner(time.Hour, time.Hour, nil)
	}, "refresh 必须大于 expired")

	a.PanicString(func() {
		NewSigner(0, 0, nil)
	}, "expired 必须大于 0")

	a.NotPanic(func() {
		NewSigner(time.Hour, 0, nil)
	})
}
