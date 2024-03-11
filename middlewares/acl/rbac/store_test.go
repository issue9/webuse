// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package rbac

import (
	"testing"

	"github.com/issue9/assert/v4"
)

var _ Store[int64] = &cacheStore[int64]{}

func newRole[T comparable](id string) *Role[T] {
	return &Role[T]{
		ID:     id + "_id",
		Parent: "",
		parent: nil,
		Name:   id + "_name",
		Desc:   id + "_desc",
	}
}

func TestCacheStore(t *testing.T) {
	a := assert.New(t, false)
	srv := newServer(a)
	s := NewCacheStore[int64](srv, "c_")
	a.NotNil(s)

	r1 := newRole[int64]("1")
	r2 := newRole[int64]("2")
	r3 := newRole[int64]("3")

	// Add

	a.NotError(s.Add(r1)).
		NotError(s.Add(r2)).
		NotError(s.Add(r3)).
		PanicString(func() { s.Add(r1) }, "角色 1_id 已经存在")

	// Exists

	exists, err := s.Exists(r3.ID)
	a.NotError(err).True(exists)
	exists, err = s.Exists("not-exists")
	a.NotError(err).False(exists)

	// Set

	r1.Name = "name"
	a.NotError(s.Set(r1))

	// Del

	a.NotError(s.Del(r3.ID))
	a.NotError(s.Del("not_exists"))
	exists, err = s.Exists(r3.ID)
	a.NotError(err).False(exists)

	// Load

	all, err := s.Load()
	a.NotError(err).
		Length(all, 2)
	a.Equal(all[r1.ID].Name, "name") // 在 set 中被修改
}
