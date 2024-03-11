// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package rbac

import (
	"net/http"
	"os"

	"github.com/issue9/assert/v4"
	"github.com/issue9/web"
	"github.com/issue9/web/server"
)

func newServer(a *assert.Assertion) web.Server {
	srv, err := server.New("test", "1.0.0", &server.Options{
		HTTPServer: &http.Server{Addr: ":8080"},
		Logs:       &server.Logs{Handler: server.NewTermHandler(os.Stderr, nil)},
	})
	a.NotError(err).NotNil(srv)
	return srv
}
