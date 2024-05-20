// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package rbac

import (
	"net/http"
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/web"
	"github.com/issue9/web/server/servertest"
	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/issue9/webuse/v7/internal/testserver"
)

func TestRBAC_NewResourceGroup(t *testing.T) {
	a := assert.New(t, false)
	s := testserver.New(a)

	rbac := New(s, NewCacheStore[string](s, "c_"), func(*web.Context) (string, web.Responser) { return "1", nil })
	a.NotNil(rbac)

	group := rbac.NewResourceGroup("id", web.Phrase("test"))
	a.NotNil(group).
		PanicString(func() {
			rbac.NewResourceGroup("id", web.Phrase("test"))
		}, "已经存在同名的资源组 id")
}

func RBAC_resourceExists(t *testing.T) {
	a := assert.New(t, false)
	s := testserver.New(a)
	rbac := New(s, NewCacheStore[string](s, "c_"), func(*web.Context) (string, web.Responser) { return "1", nil })
	a.NotNil(rbac)

	g1 := rbac.NewResourceGroup("g1", nil)
	g1.New("1", nil)
	g1.New("2", nil)
	g2 := rbac.NewResourceGroup("g2", nil)
	g2.New("3", nil)
	g2.New("4", nil)
	a.True(rbac.resourceExists(joinID(g1.id, "1"))).
		True(rbac.resourceExists(joinID(g2.id, "3"))).
		False(rbac.resourceExists(joinID(g2.id, "1")))
}

func TestResources_New(t *testing.T) {
	a := assert.New(t, false)
	s := testserver.New(a)

	rbac := New(s, NewCacheStore[string](s, "c_"), func(ctx *web.Context) (string, web.Responser) {
		q, err := ctx.Queries(true)
		if err != nil {
			return "", ctx.Error(err, "")
		}
		return q.String("id", ""), nil
	})
	a.NotNil(rbac)

	rg1, err := rbac.NewRoleGroup("g1", "1")
	a.NotError(err).NotNil(rg1)

	group := rbac.NewResourceGroup("id", web.Phrase("test"))
	a.NotNil(group)

	m1 := group.New("id1", web.Phrase("desc"))
	a.NotNil(m1)

	a.PanicString(func() {
		group.New("id1", web.Phrase("desc"))
	}, "已经存在同名的资源 id_id1")

	defer servertest.Run(a, s)()
	defer s.Close(0)

	router := s.Routers().New("def", nil)
	router.Get("/test", m1(func(*web.Context) web.Responser {
		return web.Created(nil, "")
	}, http.MethodGet, "/test"))

	// super
	servertest.Get(a, "http://localhost:8080/test?id=1").Do(nil).Status(http.StatusCreated)

	// forbidden
	servertest.Get(a, "http://localhost:8080/test?id=forbidden").Do(nil).Status(http.StatusForbidden)
}

func TestRBAC_Resources(t *testing.T) {
	a := assert.New(t, false)
	s := testserver.New(a)
	rbac := New(s, NewCacheStore[string](s, "c_"), func(*web.Context) (string, web.Responser) { return "1", nil })
	a.NotNil(rbac)

	g1 := rbac.NewResourceGroup("g1", web.Phrase("test"))
	g1.New("id1", web.Phrase("id1"))
	g1.New("id2", web.Phrase("id2"))
	g2 := rbac.NewResourceGroup("g2", web.Phrase("test"))
	g2.New("id1", web.Phrase("id1"))
	g2.New("id2", web.Phrase("id2"))

	a.Length(rbac.Resources(message.NewPrinter(language.SimplifiedChinese)), 2)
}

func TestRole_Resource(t *testing.T) {
	a := assert.New(t, false)
	s := testserver.New(a)
	rbac := New(s, NewCacheStore[string](s, "c_"), func(*web.Context) (string, web.Responser) { return "1", nil })
	a.NotNil(rbac)
	rg1, err := rbac.NewRoleGroup("g1", "")
	a.NotError(err).NotNil(rg1)

	g1 := rbac.NewResourceGroup("g1", web.Phrase("test"))
	g1.New("id1", web.Phrase("id1"))
	g1.New("id2", web.Phrase("id2"))
	g2 := rbac.NewResourceGroup("g2", web.Phrase("test"))
	g2.New("id1", web.Phrase("id1"))
	g2.New("id2", web.Phrase("id2"))

	r1, err := rg1.NewRole("r1", "r1 desc", "")
	a.NotError(err).NotNil(r1)
	r1.Allow(joinID(g1.id, "id1"), joinID(g2.id, "id2"))

	r2, err := rg1.NewRole("r2", "r2 desc", r1.ID)
	a.NotError(err).NotNil(r2)

	a.Equal(r1.Resource(), &RoleResource{
		Current: []string{joinID(g1.id, "id1"), joinID(g2.id, "id2")},
		Parent: []string{ // 没有父角色，则用全局的。
			joinID(g1.id, "id1"),
			joinID(g1.id, "id2"),
			joinID(g2.id, "id1"),
			joinID(g2.id, "id2"),
		},
	})

	a.Equal(r2.Resource(), &RoleResource{
		Parent: []string{joinID(g1.id, "id1"), joinID(g2.id, "id2")},
	})
}
