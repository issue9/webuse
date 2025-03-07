// SPDX-FileCopyrightText: 2025 caixw
//
// SPDX-License-Identifier: MIT

package systat

import (
	"bytes"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/issue9/assert/v4"
	"github.com/issue9/events"
	"github.com/issue9/logs/v7"
	"github.com/issue9/web"
	"github.com/issue9/web/mimetype/nop"
	"github.com/issue9/web/mimetype/sse"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/servertest"
)

var _ events.Subscriber[*Stats] = &service{}

func TestSystat(t *testing.T) {
	a := assert.New(t, false)
	s, err := server.NewHTTP("test", "1.0.0", &server.Options{
		HTTPServer: &http.Server{Addr: ":8080"},
		Codec:      web.NewCodec().AddMimetype(sse.Mimetype, nop.Marshal, nop.Unmarshal, ""),
		Logs:       logs.New(logs.NewTermHandler(os.Stderr, nil), logs.WithLevels(logs.AllLevels()...)),
	})
	a.NotError(err).NotNil(s)

	defer servertest.Run(a, s)()
	defer s.Close(0)

	sub := Init(s, time.Second, 10)
	o1 := &bytes.Buffer{}
	o2 := &bytes.Buffer{}

	sub.Subscribe(func(data *Stats) {
		o1.Write([]byte(data.Created.String()))
	})
	time.Sleep(2 * time.Second)
	a.NotZero(o1.Len()).
		Zero(o2.Len())

	// 后订阅，但是内容应该和之前 o1 是一样的

	sub.Subscribe(func(data *Stats) {
		o2.Write([]byte(data.Created.String()))
	})

	a.NotZero(o1.Len()).
		NotZero(o2.Len()).
		Equal(o1.String(), o2.String())
}
