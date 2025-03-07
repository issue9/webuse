// SPDX-FileCopyrightText: 2022-2024 caixw
//
// SPDX-License-Identifier: MIT

package validator

import (
	"testing"

	"github.com/issue9/assert/v4"
)

type object struct {
	Name string
	Age  int
}

func TestIn(t *testing.T) {
	a := assert.New(t, false)

	rule1 := In(1, 2)
	a.False(rule1(3))
	a.True(rule1(1))

	rule2 := In(object{}, object{Name: "name"})
	a.True(rule2(object{}))
	a.True(rule2(object{Name: "name"}))
	a.False(rule2(object{Name: "name", Age: 1}))

	rule3 := In(&object{}, &object{Name: "name"})
	a.False(rule3(&object{}))
	a.False(rule3(&object{Name: "name"}))
	a.False(rule3(&object{Name: "name", Age: 1}))
}

func TestNotIn(t *testing.T) {
	a := assert.New(t, false)

	rule1 := NotIn(1, 2)
	a.True(rule1(3))
	a.False(rule1(1))

	rule2 := NotIn(object{}, object{Name: "name"})
	a.False(rule2(object{}))
	a.False(rule2(object{Name: "name"}))
	a.True(rule2(object{Name: "name", Age: 1}))

	rule3 := NotIn(&object{}, &object{Name: "name"})
	a.True(rule3(&object{}))
	a.True(rule3(&object{Name: "name"}))
	a.True(rule3(&object{Name: "name", Age: 1}))
}
