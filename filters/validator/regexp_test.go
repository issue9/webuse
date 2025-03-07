// SPDX-FileCopyrightText: 2022-2024 caixw
//
// SPDX-License-Identifier: MIT

package validator

import (
	"testing"

	"github.com/issue9/assert/v4"
)

func TestCNPhone(t *testing.T) {
	a := assert.New(t, false)

	a.True(CNPhone("444488888888-4444"))
	a.True(CNPhone("3337777777-1"))
	a.True(CNPhone("333-7777777-1"))
	a.True(CNPhone("7777777"))
	a.True(CNPhone("88888888"))

	a.False(CNPhone("333-7777777-"))      // 尾部没有分机号
	a.False(CNPhone("22-88888888"))       // 区号只有2位
	a.False(CNPhone("33-88888888-55555")) // 分机号超过4位
}

func TestCNTel(t *testing.T) {
	a := assert.New(t, false)

	a.True(CNTel("444488888888-4444"))
	a.True(CNTel("3337777777-1"))
	a.True(CNTel("333-7777777-1"))
	a.True(CNTel("7777777"))
	a.True(CNTel("88888888"))
	a.True(CNTel("15011111111"))
	a.True(CNTel("015011111111"))
	a.True(CNTel("8615011111111"))
	a.True(CNTel("+8615011111111"))
}
