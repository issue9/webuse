// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package rbac

import (
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/web"
)

func TestRBAC_Add(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a)
	rbac, err := New(s, "", NewCacheStore[string](s, "c_"), s.Logs().INFO(), func(*web.Context) (string, web.Responser) { return "1", nil })
	a.NotError(err).NotNil(rbac)

	r1, err := rbac.Add("r1", "r1 desc", "")
	a.NotError(err).NotNil(r1).Nil(r1.parent)

	r2, err := rbac.Add("r2", "r2 desc", r1.ID)
	a.NotError(err).NotNil(r2).Equal(r2.parent, r1)

	roles, err := rbac.store.Load()
	a.NotError(err).Length(roles, 2).Equal(rbac.Role(r2.ID).parent, rbac.Role(r1.ID))

	// rbac.Roles

	maps, err := rbac.Roles()
	a.NotError(err).Length(maps, 2)

	maps, err = rbac.Roles("")
	a.NotError(err).Length(maps, 1).Equal(maps[0].ID, r1.ID)

	maps, err = rbac.Roles(r1.ID)
	a.NotError(err).Length(maps, 1).Equal(maps[0].ID, r2.ID)
}

func TestRole_Allow(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a)
	rbac, err := New(s, "", NewCacheStore[string](s, "c_"), s.Logs().INFO(), func(*web.Context) (string, web.Responser) { return "1", nil })
	a.NotError(err).NotNil(rbac)

	r1, err := rbac.Add("r1", "r1 desc", "")
	a.NotError(err).NotNil(r1).Nil(r1.parent)
	a.Equal(r1.Allow("not-exists"), web.NewLocaleError("not found resource %s", "not-exists")).
		Empty(r1.Resources)

	res11 := joinID("res1", "1")
	res12 := joinID("res1", "2")

	g := rbac.NewGroup("res1", nil)
	g.New("1", nil)
	g.New("2", nil)

	a.NotError(r1.Allow(res11)).Length(r1.Resources, 1)

	// r2 继承自 r1

	r2, err := rbac.Add("r2", "r2 desc", r1.ID)
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
	s := newServer(a)
	rbac, err := New(s, "", NewCacheStore[string](s, "c_"), s.Logs().INFO(), func(*web.Context) (string, web.Responser) { return "1", nil })
	a.NotError(err).NotNil(rbac)

	r1, err := rbac.Add("r1", "r1 desc", "")
	a.NotError(err).NotNil(r1).Nil(r1.parent)
	r2, err := rbac.Add("r2", "r2 desc", r1.ID)
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
	s := newServer(a)
	rbac, err := New(s, "", NewCacheStore[string](s, "c_"), s.Logs().INFO(), func(*web.Context) (string, web.Responser) { return "1", nil })
	a.NotError(err).NotNil(rbac)

	r1, err := rbac.Add("r1", "r1 desc", "")
	a.NotError(err).NotNil(r1).Nil(r1.parent)

	a.NotError(r1.Link("user1")).
		NotError(r1.Link("user1")).
		NotError(r1.Link("user2")).
		Equal(r1.Users, []string{"user1", "user2"}).
		NotError(r1.Unlink("user1")).
		Equal(r1.Users, []string{"user2"})
}

func TestRole_Set(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a)
	rbac, err := New(s, "", NewCacheStore[string](s, "c_"), s.Logs().INFO(), func(*web.Context) (string, web.Responser) { return "1", nil })
	a.NotError(err).NotNil(rbac)

	r1, err := rbac.Add("r1", "r1 desc", "")
	a.NotError(err).NotNil(r1).Nil(r1.parent)
	r1.Set("name", "desc")
	a.Equal(r1.Name, "name").Equal(r1.Desc, "desc")
	roles, err := rbac.store.Load()
	a.NotError(err).Length(roles, 1).
		Equal(roles[r1.ID].Name, "name").Equal(roles[r1.ID].Desc, "desc")
}

func TestRole_Roles(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a)
	rbac, err := New(s, "", NewCacheStore[string](s, "c_"), s.Logs().INFO(), func(*web.Context) (string, web.Responser) { return "1", nil })
	a.NotError(err).NotNil(rbac)

	r1, err := rbac.Add("r1", "r1 desc", "")
	a.NotError(err).NotNil(r1).Nil(r1.parent)

	r2, err := rbac.Add("r2", "r2 desc", r1.ID)
	a.NotError(err).NotNil(r2).Equal(r2.parent, r1)

	r3, err := rbac.Add("r3", "r3 desc", r2.ID)
	a.NotError(err).NotNil(r3).Equal(r3.parent, r2)

	roles, err := r1.Roles(false)
	a.NotError(err).Length(roles, 1).Equal(roles[0].ID, r2.ID)

	roles, err = r1.Roles(true)
	a.NotError(err).Length(roles, 2).
		Equal(roles[0].ID, r2.ID).
		Equal(roles[1].ID, r3.ID)
}
