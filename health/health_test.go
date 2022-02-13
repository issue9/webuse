// SPDX-License-Identifier: MIT

package health

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"
	"github.com/issue9/cache/memory"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/servertest"
)

var _ server.Middleware = &Health{}

func TestHealth(t *testing.T) {
	a := assert.New(t, false)
	srv := servertest.NewServer(a, nil)

	mem := NewCache(memory.New(1*time.Minute), "health_", log.New(os.Stderr, "[HEALTH]", 0))
	h := New(mem)
	state := mem.Get(http.MethodGet, "/")
	a.Equal(0, state.Count)

	// 第一次访问 GET /
	w := httptest.NewRecorder()
	r := rest.Get(a, "/").Request()
	h.Middleware(servertest.BuildHandler(200))(srv.NewContext(w, r))
	time.Sleep(500 * time.Millisecond) // 保存是异步的，等待完成
	state = mem.Get(http.MethodGet, "/")
	a.Equal(1, state.Count)

	// 第二次访问 GET /
	w = httptest.NewRecorder()
	r = rest.Get(a, "/").Request()
	h.Middleware(servertest.BuildHandler(500))(srv.NewContext(w, r))
	time.Sleep(500 * time.Millisecond) // 保存是异步的，等待完成
	state = mem.Get(http.MethodGet, "/")
	a.Equal(2, state.Count).
		Equal(1, state.ServerErrors).
		Equal(0, state.UserErrors)

	// 第一次访问 OPTIONS /
	w = httptest.NewRecorder()
	r = rest.NewRequest(a, http.MethodOptions, "/").Request()
	h.Middleware(servertest.BuildHandler(200))(srv.NewContext(w, r))
	time.Sleep(500 * time.Millisecond) // 保存是异步的，等待完成
	state = mem.Get(http.MethodOptions, "/")
	a.Equal(1, state.Count)

	// 第一次访问 DELETE /users
	w = httptest.NewRecorder()
	r = rest.Delete(a, "/users").Request()
	h.Middleware(servertest.BuildHandler(401))(srv.NewContext(w, r))
	time.Sleep(500 * time.Millisecond) // 保存是异步的，等待完成
	state = mem.Get(http.MethodDelete, "/users")
	a.Equal(1, state.Count).
		Equal(0, state.ServerErrors).
		Equal(1, state.UserErrors)

	all := h.States()
	a.Equal(3, len(all))
}
