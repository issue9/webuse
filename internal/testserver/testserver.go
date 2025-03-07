// SPDX-FileCopyrightText: 2025 caixw
//
// SPDX-License-Identifier: MIT

// Package testserver 提供测试用的 [web.Server] 对象
package testserver

import (
	"net/http"
	"os"

	"github.com/issue9/assert/v4"
	"github.com/issue9/logs/v7"
	"github.com/issue9/web"
	"github.com/issue9/web/mimetype/json"
	"github.com/issue9/web/server"
	"golang.org/x/text/language"
)

func New(a *assert.Assertion) web.Server {
	s, err := server.NewHTTP("test", "1.0.0", &server.Options{
		Language:   language.SimplifiedChinese,
		HTTPServer: &http.Server{Addr: ":8080"},
		Codec:      web.NewCodec().AddMimetype(json.Mimetype, json.Marshal, json.Unmarshal, json.ProblemMimetype),
		Logs:       logs.New(logs.NewTermHandler(os.Stderr, nil)),
	})
	a.NotError(err).NotNil(s)

	return s
}
