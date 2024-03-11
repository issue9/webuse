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
)

func TestRBAC_NewResources(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a)
	rbac, err := New(s, "", NewCacheStore[string](s, "c_"), s.Logs().INFO(), func(*web.Context) (string, web.Responser) { return "1", nil })
	a.NotError(err).NotNil(rbac)

	group := rbac.NewResources("id", web.Phrase("test"))
	a.NotNil(group).
		PanicString(func() {
			rbac.NewResources("id", web.Phrase("test"))
		}, "已经存在同名的资源组 id")
}

func RBAC_resourceExists(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a)
	rbac, err := New(s, "", NewCacheStore[string](s, "c_"), s.Logs().INFO(), func(*web.Context) (string, web.Responser) { return "1", nil })
	a.NotError(err).NotNil(rbac)

	g1 := rbac.NewResources("g1", nil)
	g1.New("1", nil)
	g1.New("2", nil)
	g2 := rbac.NewResources("g2", nil)
	g2.New("3", nil)
	g2.New("4", nil)
	a.True(rbac.resourceExists(g1.id + string(idSeparator) + "1")).
		True(rbac.resourceExists(g2.id + string(idSeparator) + "3")).
		False(rbac.resourceExists(g2.id + string(idSeparator) + "1"))
}

func TestResources_New(t *testing.T) {
	a := assert.New(t, false)
	s := newServer(a)

	rbac, err := New(s, "1", NewCacheStore[string](s, "c_"), s.Logs().INFO(), func(ctx *web.Context) (string, web.Responser) {
		q, err := ctx.Queries(true)
		if err != nil {
			return "", ctx.Error(err, "")
		}
		return q.String("id", ""), nil
	})
	a.NotError(err).NotNil(rbac)

	group := rbac.NewResources("id", web.Phrase("test"))
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
	}))

	// super
	servertest.Get(a, "http://localhost:8080/test?id=1").Do(nil).Status(http.StatusCreated)

	// forbidden
	servertest.Get(a, "http://localhost:8080/test?id=forbidden").Do(nil).Status(http.StatusForbidden)
}
