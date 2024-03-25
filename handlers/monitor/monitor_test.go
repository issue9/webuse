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
	"github.com/issue9/web"
	"github.com/issue9/web/mimetype/nop"
	"github.com/issue9/web/mimetype/sse"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/servertest"
)

var _ web.HandlerFunc = (&Monitor{}).Handle

func TestMonitor(t *testing.T) {
	a := assert.New(t, false)
	s, err := server.New("test", "1.0.0", &server.Options{
		HTTPServer: &http.Server{Addr: ":8080"},
		Mimetypes:  []*server.Mimetype{{Name: sse.Mimetype, Marshal: nop.Marshal, Unmarshal: nop.Unmarshal}},
		Logs:       &server.Logs{Handler: server.NewTermHandler(os.Stderr, nil), Levels: server.AllLevels()},
	})
	a.NotError(err).NotNil(s)

	defer servertest.Run(a, s)()
	defer s.Close(0)

	r := s.Routers().New("def", nil)
	m := New(s, time.Second)
	r.Get("/stats", m.Handle)

	stats := make(chan *sse.Message, 10)
	req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/stats", nil)
	a.NotError(err).NotNil(req)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = sse.OnMessage(ctx, s.Logs().ERROR(), req, nil, stats)
	a.NotError(err)

	a.Contains((<-stats).Data[0], `"cpu":`).
		Contains((<-stats).Data[0], `"mem":`)
}
