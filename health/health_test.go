// SPDX-License-Identifier: MIT

package health

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/issue9/assert/v2"
	"github.com/issue9/cache/memory"
	"github.com/issue9/web"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/servertest"
)

var _ server.Middleware = &Health{}

func TestHealth(t *testing.T) {
	a := assert.New(t, false)
	s := servertest.NewTester(a, &web.Options{Cache: memory.New(1 * time.Minute), Port: ":8080"})

	h := NewWithServer(s.Server(), "health_")
	r := s.NewRouter(h)
	r.Get("/", func(ctx *web.Context) web.Responser {
		status, err := strconv.Atoi(ctx.Request().FormValue("status"))
		if err != nil {
			panic(err)
		}
		return ctx.Status(status)
	})
	r.Post("/", func(ctx *web.Context) web.Responser {
		return nil
	})
	r.Delete("/users", func(ctx *web.Context) web.Responser {
		status, err := strconv.Atoi(ctx.Request().FormValue("status"))
		if err != nil {
			panic(err)
		}
		return ctx.Status(status)
	})

	s.GoServe()

	mem := h.store
	state := mem.Get(http.MethodGet, "/")
	a.Equal(0, state.Count)

	// 第一次访问 GET /
	s.Get("/").Query("status", "200").Do(nil).Status(200)
	state = mem.Get(http.MethodGet, "/")
	a.Equal(1, state.Count).True(state.Min > 0)

	// 第二次访问 GET /
	s.Get("/").Query("status", "500").Do(nil)
	state = mem.Get(http.MethodGet, "/")
	a.Equal(2, state.Count).
		Equal(1, state.ServerErrors).
		Equal(0, state.UserErrors)

	// 第一次访问 POST /
	s.NewRequest(http.MethodPost, "/", nil).Query("status", "201").Do(nil)
	state = mem.Get(http.MethodPost, "/")
	a.Equal(1, state.Count)

	// 第一次访问 DELETE /users
	s.Delete("/users").Query("status", "401").Do(nil)
	state = mem.Get(http.MethodDelete, "/users")
	a.Equal(1, state.Count).
		Equal(0, state.ServerErrors).
		Equal(1, state.UserErrors)

	all := h.States()
	a.Equal(3, len(all))

	s.Close(0)
	s.Wait()
}
