// SPDX-License-Identifier: MIT

package health

import (
	"math/rand"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/issue9/assert/v3"
	"github.com/issue9/web"
	"github.com/issue9/web/cache/caches"
	"github.com/issue9/web/server/servertest"
)

var _ web.Middleware = &Health{}

func TestHealth(t *testing.T) {
	a := assert.New(t, false)
	s := servertest.NewTester(a, &web.Options{Cache: caches.NewMemory(1 * time.Minute), HTTPServer: &http.Server{Addr: ":8080"}})

	h := New(NewCacheStore(s.Server(), "health_"))
	r := s.Router()
	r.Use(h)
	r.Get("/", func(ctx *web.Context) web.Responser {
		status, err := strconv.Atoi(ctx.Request().FormValue("status"))
		if err != nil {
			panic(err)
		}
		time.Sleep(time.Microsecond * time.Duration(rand.Int63n(100))) // 防止过快，无法记录用时。
		return web.Status(status)
	})
	r.Post("/", func(ctx *web.Context) web.Responser {
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

	s.GoServe()
	defer s.Close(0)

	mem := h.store
	state := mem.Get(r.Name(), http.MethodGet, "/")
	a.Equal(0, state.Count)

	// 第一次访问 GET /
	s.Get("/").Query("status", "200").Do(nil).Status(200)
	time.Sleep(time.Microsecond * 500)
	state = mem.Get(r.Name(), http.MethodGet, "/")
	a.Equal(1, state.Count).True(state.Min > 0)

	// 第二次访问 GET /
	s.Get("/").Query("status", "500").Do(nil)
	time.Sleep(time.Microsecond * 500)
	state = mem.Get(r.Name(), http.MethodGet, "/")
	a.Equal(2, state.Count).
		Equal(1, state.ServerErrors).
		Equal(0, state.UserErrors)

	// 第一次访问 POST /
	s.NewRequest(http.MethodPost, "/", nil).Query("status", "201").Do(nil)
	time.Sleep(time.Microsecond * 500)
	state = mem.Get(r.Name(), http.MethodPost, "/")
	a.Equal(1, state.Count)

	// 第一次访问 DELETE /users
	s.Delete("/users").Query("status", "401").Do(nil)
	time.Sleep(time.Microsecond * 500)
	state = mem.Get(r.Name(), http.MethodDelete, "/users")
	a.Equal(1, state.Count).
		Equal(0, state.ServerErrors).
		Equal(1, state.UserErrors)

	all := h.States()
	a.Equal(3, len(all))
}
