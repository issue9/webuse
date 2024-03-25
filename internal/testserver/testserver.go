// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

// Package testserver 提供测试用的 [web.Server] 对象
package testserver

import (
	"net/http"
	"os"

	"github.com/issue9/assert/v4"
	"github.com/issue9/web"
	"github.com/issue9/web/server"
)

func New(a *assert.Assertion) web.Server {
	s, err := server.New("test", "1.0.0", &server.Options{
		HTTPServer: &http.Server{Addr: ":8080"},
		Mimetypes:  server.JSONMimetypes(),
		Logs:       &server.Logs{Handler: server.NewTermHandler(os.Stderr, nil)},
	})
	a.NotError(err).NotNil(s)

	return s
}
