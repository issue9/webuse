// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package monitor

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/issue9/assert/v4"
	"github.com/issue9/logs/v7"
	"github.com/issue9/web"
	"github.com/issue9/web/mimetype/nop"
	"github.com/issue9/web/mimetype/sse"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/servertest"
)

var _ web.HandlerFunc = (&Monitor{}).Handle

func TestMonitor(t *testing.T) {
	a := assert.New(t, false)
	s, err := server.NewHTTP("test", "1.0.0", &server.Options{
		HTTPServer: &http.Server{Addr: ":8080"},
		Codec:      web.NewCodec().AddMimetype(sse.Mimetype, nop.Marshal, nop.Unmarshal, ""),
		Logs:       logs.New(logs.NewTermHandler(os.Stderr, nil), logs.WithLevels(logs.AllLevels()...)),
	})
	a.NotError(err).NotNil(s)

	defer servertest.Run(a, s)()
	defer s.Close(0)

	r := s.Routers().New("def", nil)
	m := New(s, time.Second)
	r.Get("/stats", m.Handle)

	stats := make(chan *sse.Message, 10)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err = sse.OnMessage(ctx, s.Logs().ERROR(), "http://localhost:8080/stats", nil, stats)
	a.When(err == nil, func(a *assert.Assertion) {
		s := <-stats
		a.Contains(s.Data[0], `"cpu":`).
			Contains(s.Data[0], `"mem":`)
	}, "返回了错误信息 %v", err)
}
