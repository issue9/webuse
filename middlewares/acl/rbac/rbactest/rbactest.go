// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

// Package rbactest 提供对 [rbac.Store] 相关的测试
package rbactest

import (
	"github.com/issue9/assert/v4"

	"github.com/issue9/webuse/v7/middlewares/acl/rbac"
)

func newRole[T comparable](id string) *rbac.Role[T] {
	return &rbac.Role[T]{
		ID:     id + "_id",
		Parent: "",
		Name:   id + "_name",
		Desc:   id + "_desc",
	}
}

// Test 执行类型为 T 的测试
func Test[T comparable](a *assert.Assertion, s rbac.Store[T]) {
	r1 := newRole[T]("1")
	r2 := newRole[T]("2")
	r3 := newRole[T]("3")

	// Add

	a.NotError(s.Add("g1", r1)).
		NotError(s.Add("g1", r2)).
		NotError(s.Add("g1", r3)).
		Panic(func() { s.Add("g1", r1) })

	// Set

	r1.Name = "name"
	a.NotError(s.Set("g1", r1))

	// Del

	a.NotError(s.Del("g1", r3.ID))
	a.NotError(s.Del("g1", "not_exists"))
	a.NotError(s.Add("g1", r3)) // 已被删除，可重新设置

	// Load

	all, err := s.Load("g1")
	a.NotError(err).
		Length(all, 3)
	a.Equal(all[r1.ID].Name, "name") // 在 set 中被修改
}
