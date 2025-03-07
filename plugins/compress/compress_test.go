// SPDX-FileCopyrightText: 2024-2025 caixw
//
// SPDX-License-Identifier: MIT

package compress

import (
	"testing"
	"time"

	"github.com/issue9/assert/v4"

	"github.com/issue9/webuse/v7/internal/testserver"
)

func TestNew(t *testing.T) {
	a := assert.New(t, false)

	a.PanicString(func() {
		New(time.Microsecond, time.Millisecond, 80)
	}, "dur 必须大于 interval")

	a.PanicString(func() {
		New(time.Millisecond, time.Microsecond, -1)
	}, "percent 必须介于 [0,100]")

	a.PanicString(func() {
		New(time.Millisecond, time.Microsecond, 101)
	}, "percent 必须介于 [0,100]")

	s := testserver.New(a)
	s.SetCompress(false)

	a.False(s.CanCompress())
	s.Use(New(time.Second, time.Microsecond, 80))
	a.Wait(time.Second * 2).True(s.CanCompress())
}
