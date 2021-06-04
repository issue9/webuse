// SPDX-License-Identifier: MIT

package compress

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/issue9/assert"
	"github.com/issue9/assert/rest"

	"github.com/issue9/middleware/v4"
)

func newCompress(a *assert.Assertion, types ...string) *Compress {
	return Default(log.New(os.Stderr, "", 0), types...)
}

func TestNew(t *testing.T) {
	a := assert.New(t)

	c := New(log.New(os.Stderr, "", log.LstdFlags), nil, "application/xml", "text/*", "application/json")
	a.NotNil(c)
	a.Equal(c.typePrefix, []string{"text/"})
	a.Equal(c.types, []string{"application/xml", "application/json"})

	a.PanicString(func() {
		New(nil, nil)
	}, "参数 errlog 不能为空")
}

var data = []*struct {
	name string

	types []string

	handler http.HandlerFunc

	// req
	reqHeaders map[string]string

	// response
	respStatus  int
	respHeaders map[string]string
	respBody    string
}{
	{
		name: "空",
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusAccepted)
		},
		respStatus: http.StatusAccepted,
	},

	{ // 在 Accept-Encoding 为空时， roundTrip 会自动处理 Accept-Encoding 为 gzip
		name:  "Content-type && WriteHeader && Write() 无 accept-encoding",
		types: []string{"text/*"},
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("content-type", "text/html")
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte("text\nhtml"))
		},
		respStatus:  http.StatusAccepted,
		respHeaders: map[string]string{"Content-Type": "text/html", "Vary": "Content-Encoding", "Content-Encoding": ""},
		respBody:    "text\nhtml",
	},

	{
		name:  "Content-type && WriteHeader && Write() accept-encoding=*",
		types: []string{"text/*"},
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("content-type", "text/html")
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte("text\nhtml"))
		},
		reqHeaders:  map[string]string{"Accept-encoding": "*"},
		respStatus:  http.StatusAccepted,
		respHeaders: map[string]string{"Content-Type": "text/html", "Vary": "Content-Encoding", "Content-Encoding": "deflate"},
		respBody:    "text\nhtml",
	},

	{
		name:  "Content-type && WriteHeader && Write() accept-encoding=identity",
		types: []string{"text/*"},
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("content-type", "text/html")
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte("text\nhtml"))
		},
		reqHeaders:  map[string]string{"Accept-encoding": "identity"},
		respStatus:  http.StatusAccepted,
		respHeaders: map[string]string{"Content-Type": "text/html", "Vary": "", "Content-Encoding": ""},
		respBody:    "text\nhtml",
	},

	{
		name:  "Content-type && WriteHeader && Write() accept-encoding=",
		types: []string{"text/*"},
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("content-type", "text/html")
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte("text\nhtml"))
		},
		reqHeaders:  map[string]string{"Accept-encoding": ""},
		respStatus:  http.StatusAccepted,
		respHeaders: map[string]string{"Content-Type": "text/html", "Vary": "", "Content-Encoding": ""},
		respBody:    "text\nhtml",
	},

	{
		name:  "Content-type && WriteHeader && Write() accept-encoding=deflate",
		types: []string{"text/*"},
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("content-type", "text/html")
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte("text\nhtml"))
		},
		reqHeaders:  map[string]string{"Accept-Encoding": "deflate"},
		respStatus:  http.StatusAccepted,
		respHeaders: map[string]string{"Content-Type": "text/html", "Vary": "Content-Encoding", "Content-Encoding": "deflate"},
		respBody:    "text\nhtml",
	},

	{
		name:  "Content-type && WriteHeader && Write() accept-encoding=gzip",
		types: []string{"text/*"},
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("content-type", "text/html")
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte("text\nhtml"))
		},
		reqHeaders:  map[string]string{"Accept-encoding": "gzip"},
		respStatus:  http.StatusAccepted,
		respHeaders: map[string]string{"Content-Type": "text/html", "Vary": "Content-Encoding", "Content-Encoding": "gzip"},
		respBody:    "text\nhtml",
	},

	{
		name:  "Content-type && WriteHeader && Write() accept-encoding=br",
		types: []string{"text/*"},
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("content-type", "text/html")
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte("text\nhtml"))
		},
		reqHeaders:  map[string]string{"Accept-encoding": "br"},
		respStatus:  http.StatusAccepted,
		respHeaders: map[string]string{"Content-Type": "text/html", "Vary": "Content-Encoding", "Content-Encoding": "br"},
		respBody:    "text\nhtml",
	},

	{
		name:  "不匹配的类型 accept-encoding=br",
		types: []string{"image/*"},
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("content-type", "text/html")
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte("text\nhtml"))
		},
		reqHeaders:  map[string]string{"Accept-encoding": "br"},
		respStatus:  http.StatusAccepted,
		respHeaders: map[string]string{"Content-Type": "text/html", "Vary": "", "Content-Encoding": ""},
		respBody:    "text\nhtml",
	},

	{
		name:  "content-type && write, accept-encodding=*",
		types: []string{"text/*"},
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("content-type", "text/html")
			w.Write([]byte("text\nhtml"))
		},
		reqHeaders:  map[string]string{"Accept-encoding": "br,*"},
		respStatus:  http.StatusOK,
		respHeaders: map[string]string{"Content-Type": "text/html", "Vary": "Content-Encoding", "Content-Encoding": "br"},
		respBody:    "text\nhtml",
	},

	{
		name:  "content-type && write, accept-encodding=identity",
		types: []string{"text/*"},
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("content-type", "text/html")
			w.Write([]byte("text\nhtml"))
		},
		reqHeaders:  map[string]string{"Accept-encoding": "identity"},
		respStatus:  http.StatusOK,
		respHeaders: map[string]string{"Content-Type": "text/html", "Vary": "", "Content-Encoding": ""},
		respBody:    "text\nhtml",
	},

	{
		name:  "content-type && write, accept-encodding=",
		types: []string{"text/*"},
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("content-type", "text/html")
			w.Write([]byte("text\nhtml"))
		},
		reqHeaders:  map[string]string{"Accept-encoding": ""},
		respStatus:  http.StatusOK,
		respHeaders: map[string]string{"Content-Type": "text/html", "Vary": "", "Content-Encoding": ""},
		respBody:    "text\nhtml",
	},

	{
		name:  "content-type && write, accept-encodding=br",
		types: []string{"text/*"},
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("content-type", "text/html")
			w.Write([]byte("text\nhtml"))
		},
		reqHeaders:  map[string]string{"Accept-encoding": "br"},
		respStatus:  http.StatusOK,
		respHeaders: map[string]string{"Content-Type": "text/html", "Vary": "Content-Encoding", "Content-Encoding": "br"},
		respBody:    "text\nhtml",
	},

	{
		name:  "content-type && write, accept-encodding=br",
		types: []string{"image/*"},
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("content-type", "text/html")
			w.Write([]byte("text\nhtml"))
		},
		reqHeaders:  map[string]string{"Accept-encoding": "br"},
		respStatus:  http.StatusOK,
		respHeaders: map[string]string{"Content-Type": "text/html", "Vary": "", "Content-Encoding": ""},
		respBody:    "text\nhtml",
	},

	{
		name:  "Write(content), accept-encodding=*",
		types: []string{"text/*"},
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("text\nhtml")) // 默认被检测为 text/plain; charset=utf-8
		},
		reqHeaders:  map[string]string{"Accept-encoding": "*"},
		respStatus:  http.StatusOK,
		respHeaders: map[string]string{"Content-Type": "text/plain; charset=utf-8", "Vary": "Content-Encoding", "Content-Encoding": "deflate"},
		respBody:    "text\nhtml",
	},

	{
		name:  "Write(content), accept-encodding=identity",
		types: []string{"text/*"},
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("text\nhtml")) // 默认被检测为 text/plain; charset=utf-8
		},
		reqHeaders:  map[string]string{"Accept-encoding": "identity"},
		respStatus:  http.StatusOK,
		respHeaders: map[string]string{"Content-Type": "text/plain; charset=utf-8", "Vary": "", "Content-Encoding": ""},
		respBody:    "text\nhtml",
	},

	{
		name:  "Write(content), accept-encodding=",
		types: []string{"text/*"},
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("text\nhtml")) // 默认被检测为 text/plain; charset=utf-8
		},
		reqHeaders:  map[string]string{"Accept-encoding": ""},
		respStatus:  http.StatusOK,
		respHeaders: map[string]string{"Content-Type": "text/plain; charset=utf-8", "Vary": "", "Content-Encoding": ""},
		respBody:    "text\nhtml",
	},

	{
		name:  "Write(content), accept-encodding=deflate",
		types: []string{"text/*"},
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("text\nhtml")) // 默认被检测为 text/plain; charset=utf-8
		},
		reqHeaders:  map[string]string{"Accept-encoding": "deflate"},
		respStatus:  http.StatusOK,
		respHeaders: map[string]string{"Content-Type": "text/plain; charset=utf-8", "Vary": "Content-Encoding", "Content-Encoding": "deflate"},
		respBody:    "text\nhtml",
	},

	{
		name:  "WriteHeader && Write(), accept-encodding=*",
		types: []string{"text/*"},
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte("text\nhtml")) // 默认被检测为 text/plain; charset=utf-8
		},
		reqHeaders:  map[string]string{"Accept-encoding": "deflate,*"},
		respStatus:  http.StatusAccepted,
		respHeaders: map[string]string{"Content-Type": "text/plain; charset=utf-8", "Vary": "Content-Encoding", "Content-Encoding": "deflate"},
		respBody:    "text\nhtml",
	},

	{
		name:  "WriteHeader && Write(), accept-encodding=identity",
		types: []string{"text/*"},
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte("text\nhtml")) // 默认被检测为 text/plain; charset=utf-8
		},
		reqHeaders:  map[string]string{"Accept-encoding": "identity"},
		respStatus:  http.StatusAccepted,
		respHeaders: map[string]string{"Content-Type": "text/plain; charset=utf-8", "Vary": "", "Content-Encoding": ""},
		respBody:    "text\nhtml",
	},

	{
		name:  "WriteHeader && Write(), accept-encodding=",
		types: []string{"text/*"},
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte("text\nhtml")) // 默认被检测为 text/plain; charset=utf-8
		},
		reqHeaders:  map[string]string{"Accept-encoding": ""},
		respStatus:  http.StatusAccepted,
		respHeaders: map[string]string{"Content-Type": "text/plain; charset=utf-8", "Vary": "", "Content-Encoding": ""},
		respBody:    "text\nhtml",
	},

	{
		name:  "WriteHeader && Write(), accept-encodding=gzip",
		types: []string{"text/plain"},
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusAccepted)
			w.Write([]byte("text\nhtml")) // 默认被检测为 text/plain; charset=utf-8
		},
		reqHeaders:  map[string]string{"Accept-encoding": "deflate;q=0.9,gzip"},
		respStatus:  http.StatusAccepted,
		respHeaders: map[string]string{"Content-Type": "text/plain; charset=utf-8", "Vary": "Content-Encoding", "Content-Encoding": "gzip"},
		respBody:    "text\nhtml",
	},

	{
		name:  "WriteHeader(204) && Write(nil), accept-encodding=*",
		types: []string{"text/*"},
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
			w.Write(nil) // 默认被检测为 text/plain; charset=utf-8
		},
		reqHeaders:  map[string]string{"Accept-encoding": "*"},
		respStatus:  http.StatusNoContent,
		respHeaders: map[string]string{"Content-Type": "text/plain; charset=utf-8", "Vary": "", "Content-Encoding": ""},
	},

	{
		name:  "WriteHeader(204) && Write(nil), accept-encodding=identity",
		types: []string{"text/*"},
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
			w.Write(nil) // identity 不压缩，不修改，且始终不会 406
		},
		reqHeaders:  map[string]string{"Accept-encoding": "identity"},
		respStatus:  http.StatusNoContent,
		respHeaders: map[string]string{"Content-Type": "", "Vary": "", "Content-Encoding": ""},
	},

	{
		name:  "WriteHeader(204) && Write(nil), accept-encodding=",
		types: []string{"text/*"},
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
			w.Write(nil)
		},
		reqHeaders:  map[string]string{"Accept-encoding": ""},
		respStatus:  http.StatusNoContent,
		respHeaders: map[string]string{"Content-Type": "", "Vary": "", "Content-Encoding": ""},
	},

	{
		name:  "WriteHeader(204) && Write(nil), accept-encodding=gzip",
		types: []string{"text/*"},
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
			w.Write(nil) // 默认被检测为 text/plain; charset=utf-8
		},
		reqHeaders:  map[string]string{"Accept-encoding": "gzip,deflate;q=0.8"},
		respStatus:  http.StatusNoContent,
		respHeaders: map[string]string{"Content-Type": "text/plain; charset=utf-8", "Vary": "", "Content-Encoding": ""},
	},

	{
		name:  "WriteHeader && Write(nil), accept-encodding=*",
		types: []string{"text/*"},
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusAccepted)
			w.Write(nil) // 默认被检测为 text/plain; charset=utf-8
		},
		reqHeaders:  map[string]string{"Accept-encoding": "br,*,deflate;q=0.9"},
		respStatus:  http.StatusAccepted,
		respHeaders: map[string]string{"Content-Type": "text/plain; charset=utf-8", "Vary": "Content-Encoding", "Content-Encoding": "br"},
	},

	{
		name:  "WriteHeader && Write(nil), accept-encodding=identity",
		types: []string{"text/*"},
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusAccepted)
			w.Write(nil)
		},
		reqHeaders:  map[string]string{"Accept-encoding": "identity"},
		respStatus:  http.StatusAccepted,
		respHeaders: map[string]string{"Content-Type": "", "Vary": "", "Content-Encoding": ""},
	},

	{
		name:  "WriteHeader && Write(nil), accept-encodding=",
		types: []string{"text/*"},
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusAccepted)
			w.Write(nil)
		},
		reqHeaders:  map[string]string{"Accept-encoding": ""},
		respStatus:  http.StatusAccepted,
		respHeaders: map[string]string{"Content-Type": "", "Vary": "", "Content-Encoding": ""},
	},

	{
		name:  "WriteHeader && Write(nil), accept-encodding=gzip",
		types: []string{"text/*"},
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusAccepted)
			w.Write(nil) // 默认被检测为 text/plain; charset=utf-8
		},
		reqHeaders:  map[string]string{"Accept-encoding": "gzip"},
		respStatus:  http.StatusAccepted,
		respHeaders: map[string]string{"Content-Type": "text/plain; charset=utf-8", "Vary": "Content-Encoding", "Content-Encoding": "gzip"},
	},

	{
		name:  "Write(nil), accept-encodding=*",
		types: []string{"text/*"},
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.Write(nil) // 默认被检测为 text/plain; charset=utf-8
		},
		reqHeaders:  map[string]string{"Accept-encoding": "*"},
		respStatus:  http.StatusOK,
		respHeaders: map[string]string{"Content-Type": "text/plain; charset=utf-8", "Vary": "Content-Encoding", "Content-Encoding": "deflate"},
	},

	{
		name:  "Write(nil), accept-encodding=identity",
		types: []string{"text/*"},
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.Write(nil)
		},
		reqHeaders:  map[string]string{"Accept-encoding": "identity"},
		respStatus:  http.StatusOK,
		respHeaders: map[string]string{"Content-Type": "", "Vary": "", "Content-Encoding": ""},
	},

	{
		name:  "Write(nil), accept-encodding=",
		types: []string{"text/*"},
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.Write(nil)
		},
		reqHeaders:  map[string]string{"Accept-encoding": ""},
		respStatus:  http.StatusOK,
		respHeaders: map[string]string{"Content-Type": "", "Vary": "", "Content-Encoding": ""},
	},

	{
		name:  "Write(nil), accept-encodding=gzip",
		types: []string{"text/*"},
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.Write(nil) // 默认被检测为 text/plain; charset=utf-8
		},
		reqHeaders:  map[string]string{"Accept-encoding": "gzip"},
		respStatus:  http.StatusOK,
		respHeaders: map[string]string{"Content-Type": "text/plain; charset=utf-8", "Vary": "Content-Encoding", "Content-Encoding": "gzip"},
	},

	{
		name:  "多次调用 Write, accept-encodding=*",
		types: []string{"text/*"},
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("text")) // 默认被检测为 text/plain; charset=utf-8
			w.Write([]byte("/html"))
		},
		reqHeaders:  map[string]string{"Accept-encoding": "*"},
		respStatus:  http.StatusOK,
		respHeaders: map[string]string{"Content-Type": "text/plain; charset=utf-8", "Vary": "Content-Encoding", "Content-Encoding": "deflate"},
		respBody:    "text/html",
	},

	{
		name:  "多次调用 Write, accept-encodding=identity",
		types: []string{"text/*"},
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("text")) // 默认被检测为 text/plain; charset=utf-8
			w.Write([]byte("/html"))
		},
		reqHeaders:  map[string]string{"Accept-encoding": "identity"},
		respStatus:  http.StatusOK,
		respHeaders: map[string]string{"Content-Type": "text/plain; charset=utf-8", "Vary": "", "Content-Encoding": ""},
		respBody:    "text/html",
	},

	{
		name:  "多次调用 Write, accept-encodding=",
		types: []string{"text/*"},
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("text")) // 默认被检测为 text/plain; charset=utf-8
			w.Write([]byte("/html"))
		},
		reqHeaders:  map[string]string{"Accept-encoding": ""},
		respStatus:  http.StatusOK,
		respHeaders: map[string]string{"Content-Type": "text/plain; charset=utf-8", "Vary": "", "Content-Encoding": ""},
		respBody:    "text/html",
	},

	{
		name:  "多次调用 Write, accept-encodding=deflate",
		types: []string{"text/*"},
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("text")) // 默认被检测为 text/plain; charset=utf-8
			w.Write([]byte("/html"))
		},
		reqHeaders:  map[string]string{"Accept-encoding": "deflate"},
		respStatus:  http.StatusOK,
		respHeaders: map[string]string{"Content-Type": "text/plain; charset=utf-8", "Vary": "Content-Encoding", "Content-Encoding": "deflate"},
		respBody:    "text/html",
	},

	{
		name:  "write(nil) && Write(content), accept-encodding=deflate",
		types: []string{"text/*"},
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.Write(nil) // 默认被检测为 text/plain; charset=utf-8
			w.Write([]byte("/html"))
		},
		reqHeaders:  map[string]string{"Accept-encoding": "deflate"},
		respStatus:  http.StatusOK,
		respHeaders: map[string]string{"Content-Type": "text/plain; charset=utf-8", "Vary": "Content-Encoding", "Content-Encoding": "deflate"},
		respBody:    "/html",
	},

	{
		name:  "writeHeader(204) && Write(content), accept-encodding=deflate",
		types: []string{"text/*"},
		handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
			w.Write(nil) // 默认被检测为 text/plain; charset=utf-8
			w.Write([]byte("/html"))
		},
		reqHeaders:  map[string]string{"Accept-encoding": "deflate"},
		respStatus:  http.StatusNoContent,
		respHeaders: map[string]string{"Content-Type": "text/plain; charset=utf-8", "Vary": "", "Content-Encoding": ""},
		respBody:    "",
	},
}

func TestCompress_MiddlewareFunc(t *testing.T) {
	a := assert.New(t)
	buf := new(bytes.Buffer)

	for index, item := range data {
		c := newCompress(a, item.types...)
		a.NotNil(c, "构建 Compress 对象出错，%d,%s", index, item.name)

		srv := rest.NewServer(t, c.MiddlewareFunc(item.handler), &http.Client{})
		defer srv.Close()

		// req
		req := srv.NewRequest(http.MethodGet, "/")
		for k, v := range item.reqHeaders {
			req.Header(k, v)
		}

		// resp
		resp := req.Do()
		resp.Status(item.respStatus)
		for k, v := range item.respHeaders {
			resp.Header(k, v, "返回报头[%s:%s]错误，位于:%d,%s ", k, v, index, item.name)
		}

		// resp body
		buf.Reset()
		resp.ReadBody(buf)
		var reader io.Reader
		var err error
		switch item.respHeaders["Content-Encoding"] {
		case "br":
			reader = brotli.NewReader(buf)
		case "deflate":
			reader = flate.NewReader(buf)
		case "gzip":
			reader, err = gzip.NewReader(buf)
		default:
			name := item.respHeaders["Content-Encoding"]
			a.Empty(name, "Content-Encoding 不为空 %s,位于:%d,%s", name, index, item.name)
			reader = buf
		}
		a.NotError(err).NotNil(reader)

		data, err := ioutil.ReadAll(reader)
		a.NotError(err, "%s 位于 %d:%s", err, index, item.name).NotNil(data)
		a.Equal(string(data), item.respBody)
	}
}

func TestCompress_canCompressed(t *testing.T) {
	a := assert.New(t)

	c := New(log.New(os.Stderr, "", 0), nil)
	a.NotNil(c)

	a.False(c.canCompressed(""))

	c = newCompress(a, "text/*", "application/json")

	// 未指定 content-type
	a.False(c.canCompressed(""))

	a.True(c.canCompressed("text/html;charset=utf-8"))

	a.True(c.canCompressed("application/json"))

	a.False(c.canCompressed("application/octet"))
}

// 在任何输出中间件之前应用了压缩中间件
func TestCompress_Middleware_Before(t *testing.T) {
	a := assert.New(t)

	f201 := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated) // 在中间件中提早输出了内容，此处应该不启作用。
		_, err := w.Write([]byte("201"))
		a.NotError(err)
	}
	m := middleware.NewMiddlewares(http.HandlerFunc(f201))
	a.NotNil(m)

	m.Prepend(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte("after"))
			a.NotError(err)
			h.ServeHTTP(w, r)
		})
	})
	c := New(log.New(os.Stderr, "", 0), nil, "*")
	a.NotNil(c)
	a.NotError(c.AddAlgorithm("gzip", NewGzip))
	a.NotError(c.AddAlgorithm("deflate", NewDeflate))
	m.Prepend(c.Middleware) // 插到之前

	// 未请求压缩
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/path", nil)
	m.ServeHTTP(w, r)
	a.Equal(w.Body.String(), "after201")
	a.Equal(w.Header().Get("Content-Encoding"), "")
	a.Equal(http.StatusOK, w.Result().StatusCode)

	// 请求压缩
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/path", nil)
	r.Header.Set("Accept-Encoding", "deflate;q=0.8")

	m.ServeHTTP(w, r)
	a.Equal(w.Header().Get("Content-Encoding"), "deflate")
	a.Equal(w.Header().Get("Content-Type"), "text/plain; charset=utf-8")
	a.Equal(http.StatusOK, w.Result().StatusCode)
	reader := flate.NewReader(w.Body)
	data, err := ioutil.ReadAll(reader)
	a.NotError(err).NotNil(data)
	a.Equal(string(data), "after201")
}

func TestCompress_isIgnore(t *testing.T) {
	a := assert.New(t)

	c := New(log.New(os.Stderr, "", 0), []string{http.MethodDelete, http.MethodOptions})
	a.True(c.isIgnore(http.MethodDelete)).
		True(c.isIgnore(http.MethodOptions)).
		False(c.isIgnore(http.MethodPost))
}
