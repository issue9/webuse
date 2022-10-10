// SPDX-License-Identifier: MIT

package requestid

import (
	"net/http"
	"testing"

	"github.com/issue9/assert/v3"
	"github.com/issue9/web"
	"github.com/issue9/web/server/servertest"
)

func TestRequestID(t *testing.T) {
	a := assert.New(t, false)
	srv := servertest.NewTester(a, nil)

	srv.GoServe()
	defer srv.Close(0)

	r := srv.Router()
	r.Use(New("", nil))

	r.Get("/id1", func(ctx *web.Context) web.Responser {
		a.Equal(Get(ctx), "id1")
		return web.OK(nil)
	})

	r.Get("/gen", func(ctx *web.Context) web.Responser {
		a.NotEmpty(Get(ctx))
		return web.OK(nil)
	})

	srv.Get("/id1").
		Header("X-request-id", "id1").
		Do(nil).
		Status(http.StatusOK)

	srv.Get("/gen"). // 未设置 id
				Do(nil).
				Status(http.StatusOK)
}
