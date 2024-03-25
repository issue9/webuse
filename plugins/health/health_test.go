// SPDX-FileCopyrightText: 2015-2024 caixw
//
// SPDX-License-Identifier: MIT

package health

import (
	"math/rand"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/issue9/assert/v4"
	"github.com/issue9/web"
	"github.com/issue9/web/server/servertest"

	"github.com/issue9/webuse/v7/internal/testserver"
)

var _ web.Plugin = &Health{}

func TestHealth(t *testing.T) {
	a := assert.New(t, false)

	s := testserver.New(a)
	//dr :=s.Cache().(cache.Driver)

	h := New(NewCacheStore(s, "health_"))
	s.Use(h)
	r := s.Routers().New("def", nil)
	r.Get("/", func(ctx *web.Context) web.Responser {
		status, err := strconv.Atoi(ctx.Request().FormValue("status"))
		if err != nil {
			panic(err)
		}
		time.Sleep(time.Microsecond * time.Duration(rand.Int63n(100))) // 防止过快，无法记录用时。
		return web.Status(status)
	})
	r.Post("/", func(*web.Context) web.Responser {
		time.Sleep(time.Microsecond * time.Duration(rand.Int63n(100))) // 防止过快，无法记录用时。
		return nil
	})
	r.Delete("/users", func(ctx *web.Context) web.Responser {
		status, err := strconv.Atoi(ctx.Request().FormValue("status"))
		if err != nil {
			panic(err)
		}
		time.Sleep(time.Microsecond * time.Duration(rand.Int63n(100))) // 防止过快，无法记录用时。
		return web.Status(status)
	})

	defer servertest.Run(a, s)()
	defer s.Close(0)

	mem := h.store
	state := mem.Get(r.Name(), http.MethodGet, "/")
	a.Equal(0, state.Count)

	// 第一次访问 GET /
	servertest.Get(a, "http://localhost:8080/").Query("status", "200").Do(nil).Status(200)
	time.Sleep(time.Microsecond * 500)
	state = mem.Get(r.Name(), http.MethodGet, "/")
	a.Equal(1, state.Count).True(state.Min > 0)

	// 第二次访问 GET /
	servertest.Get(a, "http://localhost:8080/").Query("status", "500").Do(nil)
	time.Sleep(time.Microsecond * 500)
	state = mem.Get(r.Name(), http.MethodGet, "/")
	a.Equal(2, state.Count).
		Equal(1, state.ServerErrors).
		Equal(0, state.UserErrors)

	// 第一次访问 POST /
	servertest.NewRequest(a, http.MethodPost, "http://localhost:8080/").Query("status", "201").Do(nil)
	time.Sleep(time.Microsecond * 500)
	state = mem.Get(r.Name(), http.MethodPost, "/")
	a.Equal(1, state.Count)

	// 第一次访问 DELETE /users
	servertest.Delete(a, "http://localhost:8080/users").Query("status", "401").Do(nil)
	time.Sleep(time.Microsecond * 500)
	state = mem.Get(r.Name(), http.MethodDelete, "/users")
	a.Equal(1, state.Count).
		Equal(0, state.ServerErrors).
		Equal(1, state.UserErrors).
		Length(h.States(), 3)
}
