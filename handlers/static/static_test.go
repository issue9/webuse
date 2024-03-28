// SPDX-FileCopyrightText: 2018-2024 caixw
//
// SPDX-License-Identifier: MIT

package static

import (
	"bytes"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/issue9/assert/v4"
	"github.com/issue9/web/server/servertest"

	"github.com/issue9/webuse/v7/internal/testserver"
)

func TestServeFileHandler(t *testing.T) {
	a := assert.New(t, false)
	srv := testserver.New(a)
	router := srv.Routers().New("def", nil)

	defer servertest.Run(a, srv)()
	defer srv.Close(0)

	a.PanicString(func() {
		ServeFileHandler(nil, "path", "index.html")
	}, "参数 fsys 不能为空")

	a.PanicString(func() {
		ServeFileHandler(os.DirFS("./testdata"), "", "index.html")
	}, "参数 name 不能为空")

	router.Get("/serve/{path}", ServeFileHandler(os.DirFS("./testdata"), "path", "index.html"))
	servertest.Get(a, "http://localhost:8080/serve/file1.txt"). // file1.txt
									Do(nil).
									Status(http.StatusOK).
									StringBody("file1")
	servertest.Get(a, "http://localhost:8080/serve/"). // index.html
								Do(nil).
								Status(http.StatusOK).
								BodyFunc(func(a *assert.Assertion, body []byte) {
			a.True(bytes.HasPrefix(body, []byte("<!DOCTYPE html>")))
		})
}

func TestAttachmentFileHandler(t *testing.T) {
	a := assert.New(t, false)
	srv := testserver.New(a)
	router := srv.Routers().New("def", nil)

	defer servertest.Run(a, srv)()
	defer srv.Close(0)

	a.PanicString(func() {
		AttachmentFileHandler(nil, "path", "filename", true)
	}, "参数 fsys 不能为空")

	a.PanicString(func() {
		AttachmentFileHandler(os.DirFS("./testdata"), "", "filename", true)
	}, "参数 name 不能为空")

	router.Get("/attach/{path}", AttachmentFileHandler(os.DirFS("./testdata"), "path", "中文", true))

	servertest.Get(a, "http://localhost:8080/attach/file1.txt").
		Do(nil).
		Status(http.StatusOK).
		Header(contentDisposition, "inline; filename="+url.QueryEscape("中文"))
}

func TestAttachmentReaderHandler(t *testing.T) {
	a := assert.New(t, false)
	srv := testserver.New(a)
	router := srv.Routers().New("def", nil)

	defer servertest.Run(a, srv)()
	defer srv.Close(0)

	reader := bytes.NewReader([]byte("abc"))
	router.Get("/attach/path", AttachmentReaderHandler("中文", true, time.Now(), reader))

	servertest.Get(a, "http://localhost:8080/attach/path").
		Do(nil).
		Status(http.StatusOK).
		Header(contentDisposition, "inline; filename="+url.QueryEscape("中文"))
}
