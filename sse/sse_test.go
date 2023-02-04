// SPDX-License-Identifier: MIT

package sse

import (
	"strconv"
	"testing"
	"time"

	"github.com/issue9/assert/v3"
	"github.com/issue9/web"
	"github.com/issue9/web/server/servertest"
)

func TestEvents(t *testing.T) {
	a := assert.New(t, false)
	e := NewServer[int64](201)
	a.NotNil(e)
	s := servertest.NewTester(a, nil)
	s.Server().Mimetypes().Add("text/event-stream", nil, nil, "")
	s.Server().Services().Add(web.Phrase("sse"), e)

	s.Router().Get("/events/{id}", func(ctx *web.Context) web.Responser {
		id, resp := ctx.ParamInt64("id", web.ProblemBadRequest)
		if resp != nil {
			return resp
		}

		s, wait := e.NewSource(id, ctx)
		s.Sent([]string{"connect", strconv.FormatInt(id, 10)}, "", "1", 50)
		time.Sleep(time.Microsecond * 500)
		s.Sent([]string{"msg", strconv.FormatInt(id, 10)}, "event", "2", 0)
		s.Sent([]string{"msg", strconv.FormatInt(id, 10)}, "event", "2", 0)

		wait()
		return nil
	})

	s.GoServe()

	time.AfterFunc(5000*time.Microsecond, func() {
		e.Get(5).Close()
	})

	s.Get("/events/5").
		Header("accept", "text/event-stream").
		Header("accept-encoding", "").
		Do(nil).
		Status(201).
		StringBody(`data: connect
data: 5
id: 1
retry: 50

data: msg
data: 5
event: event
id: 2

data: msg
data: 5
event: event
id: 2

`)

	s.Close(0)
}
