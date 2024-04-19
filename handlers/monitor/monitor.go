// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

// Package monitor 系统状态检测
package monitor

import (
	"context"
	"encoding/json"
	"time"

	"github.com/issue9/events"
	"github.com/issue9/web"
	"github.com/issue9/web/mimetype/sse"
)

// Monitor 系统状态检测
type Monitor struct {
	dur    time.Duration
	server *sse.Server[string]
	event  *events.Event[*Stats]
	enable bool
}

// New 声明 [Monitor]
//
// dur 为发送状态数据的时间间隔；
func New(s web.Server, dur time.Duration) *Monitor {
	m := &Monitor{
		dur:    dur,
		server: sse.NewServer[string](s, 0, 0, 10),
		event:  events.New[*Stats](),
		enable: true,
	}

	s.Services().AddTicker(web.Phrase("monitor system stat"), func(now time.Time) error {
		if !m.enable {
			return nil
		}

		stats, err := calcState(dur, now)
		if err != nil {
			return err
		}

		m.event.Publish(true, stats)
		return nil
	}, dur, true, false)

	return m
}

// Handle 输出监视信息
//
// NOTE: 这是一个 SSE 连接，需要保证 content-type 的正确性。
func (m *Monitor) Handle(ctx *web.Context) web.Responser {
	source, wait := m.server.NewSource(ctx.Server().UniqueID(), ctx) // 只要保证是唯一 ID 就行
	var cancel context.CancelFunc

	defer func() {
		wait()
		cancel() // 在 wait 之后
		m.enable = m.event.Len() > 0
	}()

	event := source.NewEvent("monitor", json.Marshal)
	cancel = m.event.Subscribe(func(data *Stats) {
		if err := event.Sent(data); err != nil {
			ctx.Logs().ERROR().Error(err)
		}
	})

	return nil
}
