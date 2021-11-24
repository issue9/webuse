// SPDX-License-Identifier: MIT

package debugger

import (
	"net/http"
	"testing"

	"github.com/issue9/assert/v2"
	"github.com/issue9/assert/v2/rest"
)

var f201 = func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusCreated)
	_, err := w.Write([]byte("1234567890"))
	if err != nil {
		println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func TestDebugger_MiddlewareFunc(t *testing.T) {
	a := assert.New(t, false)
	dbg := &Debugger{}
	srv := rest.NewServer(a, dbg.MiddlewareFunc(f201), nil)

	// pprof == ""
	srv.Get("/debug/pprof/").Do(nil).Status(http.StatusCreated)

	dbg.Pprof = "/debug/pprof1/"

	srv.Get("/debug/pprof/").Do(nil).Status(http.StatusCreated) // 访问默认的 f201
	srv.Get("/debug/pprof1/").Do(nil).Status(http.StatusOK)
	srv.Get("/debug/pprof1/heap").Do(nil).Status(http.StatusOK)
	srv.Get("/debug/pprof1/cmdline").Do(nil).Status(http.StatusOK)
	srv.Get("/debug/pprof1/trace").Do(nil).Status(http.StatusOK)
	srv.Get("/debug/pprof1/symbol").Do(nil).Status(http.StatusOK)
	// srv.Get("/debug/pprof1/profile").Do().Status(http.StatusOK) // 时间较长，不测试

	// vars == ""，则访问 f201
	srv.Get("/debug/vars").Do(nil).Status(http.StatusCreated)

	// vars == /debug/vars1
	dbg.Vars = "/debug/vars1"
	srv.Get("/debug/vars").Do(nil).Status(http.StatusCreated)
	srv.Get("/debug/vars1").Do(nil).Status(http.StatusOK)

	// 命中 h201
	srv.Get("/debug/").Do(nil).Status(http.StatusCreated)
}
