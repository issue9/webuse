// SPDX-License-Identifier: MIT

package debugger

import (
	"net/http"
	"testing"

	"github.com/issue9/assert/rest"
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
	dbg := &Debugger{}
	srv := rest.NewServer(t, dbg.MiddlewareFunc(f201), nil)
	defer srv.Close()

	// pprof == ""
	srv.Get("/debug/pprof/").Do().Status(http.StatusCreated)

	dbg.Pprof = "/debug/pprof1/"

	srv.Get("/debug/pprof/").Do().Status(http.StatusCreated) // 访问默认的 f201
	srv.Get("/debug/pprof1/").Do().Status(http.StatusOK)
	srv.Get("/debug/pprof1/heap").Do().Status(http.StatusOK)
	srv.Get("/debug/pprof1/cmdline").Do().Status(http.StatusOK)
	srv.Get("/debug/pprof1/trace").Do().Status(http.StatusOK)
	srv.Get("/debug/pprof1/symbol").Do().Status(http.StatusOK)
	// srv.Get("/debug/pprof1/profile").Do().Status(http.StatusOK) // 时间较长，不测试

	// vars == ""，则访问 f201
	srv.Get("/debug/vars").Do().Status(http.StatusCreated)

	// vars == /debug/vars1
	dbg.Vars = "/debug/vars1"
	srv.Get("/debug/vars").Do().Status(http.StatusCreated)
	srv.Get("/debug/vars1").Do().Status(http.StatusOK)

	// 命中 h201
	srv.Get("/debug/").Do().Status(http.StatusCreated)
}
