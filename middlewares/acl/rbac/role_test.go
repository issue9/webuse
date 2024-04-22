// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package rbac

import (
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/web"

	"github.com/issue9/webuse/v7/internal/testserver"
)

func TestRoleGroup_New(t *testing.T) {
	a := assert.New(t, false)
	s := testserver.New(a)
	rbac := New(s, NewCacheStore[string](s, "c_"), func(*web.Context) (string, web.Responser) { return "1", nil })
	a.NotNil(rbac)
	g, err := rbac.NewRoleGroup("g1", "")
	a.NotError(err).NotNil(g)

	r1, err := g.NewRole("r1", "r1 desc", "")
	a.NotError(err).NotNil(r1).Nil(r1.parent)

	r2, err := g.NewRole("r2", "r2 desc", r1.ID)
	a.NotError(err).NotNil(r2).Equal(r2.parent, r1)

	roles, err := g.rbac.store.Load(g.id)
	a.NotError(err).Length(roles, 2).Equal(g.Role(r2.ID).parent, g.Role(r1.ID))

	// rbac.Roles

	a.Length(g.Roles(), 2)
}

func TestRole_Allow(t *testing.T) {
	a := assert.New(t, false)
	s := testserver.New(a)
	rbac := New(s, NewCacheStore[string](s, "c_"), func(*web.Context) (string, web.Responser) { return "1", nil })
	a.NotNil(rbac)
	rg1, err := rbac.NewRoleGroup("g1", "")
	a.NotError(err).NotNil(rg1)

	r1, err := rg1.NewRole("r1", "r1 desc", "")
	a.NotError(err).NotNil(r1).Nil(r1.parent)
	a.Equal(r1.Allow("not-exists"), web.NewLocaleError("not found resource %s", "not-exists")).
		Empty(r1.Resources)

	res11 := joinID("res1", "1")
	res12 := joinID("res1", "2")

	g := rbac.NewResourceGroup("res1", nil)
	g.New("1", nil)
	g.New("2", nil)

	a.NotError(r1.Allow(res11)).
		NotError(r1.Link("u1")).
		Length(r1.Resources, 1).
		True(rg1.isAllow("u1", res11))

	// r2 继承自 r1

	r2, err := rg1.NewRole("r2", "r2 desc", r1.ID)
	a.NotError(err).NotNil(r2).
		NotError(r2.Allow(res11))

	// 子角色还有 res1_1
	a.Equal(r1.Allow(res12), web.NewLocaleError("child role has resource %s can not be deleted", res11))
	a.NotError(r2.Allow())                                                       // 清空 r2 的资源
	a.NotError(r1.Allow(res12))                                                  // 现在可以改变 r1 的资源
	a.Equal(r2.Allow(res11), web.NewLocaleError("not found resource %s", res11)) // 父类 r1 不拥有 res11
}

func TestRole_Del(t *testing.T) {
	a := assert.New(t, false)
	s := testserver.New(a)
	rbac := New(s, NewCacheStore[string](s, "c_"), func(*web.Context) (string, web.Responser) { return "1", nil })
	a.NotNil(rbac)
	rg1, err := rbac.NewRoleGroup("g1", "")
	a.NotError(err).NotNil(rg1)

	r1, err := rg1.NewRole("r1", "r1 desc", "")
	a.NotError(err).NotNil(r1).Nil(r1.parent)
	r2, err := rg1.NewRole("r2", "r2 desc", r1.ID)
	a.NotError(err).NotNil(r2).Equal(r2.parent, r1)
	r2.Users = []string{"1"}

	a.Equal(r1.Del(), web.NewLocaleError("the role %s has children role, can not deleted", r1.ID))
	a.Equal(r2.Del(), web.NewLocaleError("the role %s has users, can not deleted", r2.ID))

	r2.Users = []string{}
	a.NotError(r2.Del()).
		NotError(r1.Del())
}

func TestRole_Link(t *testing.T) {
	a := assert.New(t, false)
	s := testserver.New(a)
	rbac := New(s, NewCacheStore[string](s, "c_"), func(*web.Context) (string, web.Responser) { return "1", nil })
	a.NotNil(rbac)
	rg1, err := rbac.NewRoleGroup("g1", "")
	a.NotError(err).NotNil(rg1)

	r1, err := rg1.NewRole("r1", "r1 desc", "")
	a.NotError(err).NotNil(r1).Nil(r1.parent)

	a.NotError(r1.Link("user1")).
		NotError(r1.Link("user1")).
		NotError(r1.Link("user2")).
		Equal(r1.Users, []string{"user1", "user2"}).
		Equal(rg1.UserRoles("user1"), []*Role[string]{r1}).
		NotError(r1.Unlink("user1")).
		Equal(r1.Users, []string{"user2"})

	r2, err := rg1.NewRole("r2", "r2 desc", r1.ID)
	a.NotError(err).NotNil(r1).
		Equal(r2.parent, r1).
		Equal(r2.Link("user2"), web.NewLocaleError("user %v in the parent role %s", "user2", r1.ID)).
		NotError(r2.Link("user1")) // user1 已经在上面 Unlink
}

func TestRole_Set(t *testing.T) {
	a := assert.New(t, false)
	s := testserver.New(a)
	rbac := New(s, NewCacheStore[string](s, "c_"), func(*web.Context) (string, web.Responser) { return "1", nil })
	a.NotNil(rbac)
	rg1, err := rbac.NewRoleGroup("g1", "")
	a.NotError(err).NotNil(rg1)

	r1, err := rg1.NewRole("r1", "r1 desc", "")
	a.NotError(err).NotNil(r1).Nil(r1.parent)
	r1.Set("name", "desc")
	a.Equal(r1.Name, "name").Equal(r1.Desc, "desc")
	roles, err := rg1.rbac.store.Load(rg1.id)
	a.NotError(err).Length(roles, 1).
		Equal(roles[r1.ID].Name, "name").Equal(roles[r1.ID].Desc, "desc")
}

func TestRole_Descendants(t *testing.T) {
	a := assert.New(t, false)
	s := testserver.New(a)
	rbac := New(s, NewCacheStore[string](s, "c_"), func(*web.Context) (string, web.Responser) { return "1", nil })
	a.NotNil(rbac)
	rg1, err := rbac.NewRoleGroup("g1", "")
	a.NotError(err).NotNil(rg1)

	r1, err := rg1.NewRole("r1", "r1 desc", "")
	a.NotError(err).NotNil(r1).Nil(r1.parent)

	r2, err := rg1.NewRole("r2", "r2 desc", r1.ID)
	a.NotError(err).NotNil(r2).Equal(r2.parent, r1)

	r3, err := rg1.NewRole("r3", "r3 desc", r2.ID)
	a.NotError(err).NotNil(r3).Equal(r3.parent, r2)

	roles, err := r1.Descendants(false)
	a.NotError(err).
		Length(roles, 1).
		Equal(roles[0].ID, r2.ID).
		True(r1.IsDescendant(r2.ID)).
		True(r1.IsDescendant(r3.ID)).
		True(r2.IsDescendant(r3.ID))

	roles, err = r1.Descendants(true)
	a.NotError(err).Length(roles, 2).
		Equal(roles[0].ID, r2.ID).
		Equal(roles[1].ID, r3.ID)
}
