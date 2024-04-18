// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

// Package monitor 系统状态检测
package monitor

import (
	"context"
	"encoding/json"
	"time"

	"github.com/issue9/web"
	"github.com/issue9/web/mimetype/sse"
)

// Monitor 系统状态检测
type Monitor struct {
	dur time.Duration
	s   *sse.Server[int]
}

// New 声明 [Monitor]
//
// dur 为发送状态数据的时间间隔；
func New(s web.Server, dur time.Duration) *Monitor {
	return &Monitor{
		dur: dur,
		s:   sse.NewServer[int](s, 0, 0, 10),
	}
}

// Handle 输出监视信息
//
// NOTE: 这是一个 SSE 连接，需要保证 content-type 的正确性。
func (m *Monitor) Handle(ctx *web.Context) web.Responser {
	source, wait := m.s.NewSource(0, ctx)
	event := source.NewEvent("monitor", json.Marshal)
	var cancel context.CancelFunc

	defer func() {
		wait()

		if cancel != nil { // 退出时删除 ticker 事件
			cancel()
		}
	}()

	cancel = ctx.Server().Services().AddTicker(web.Phrase("monitor system stats"), func(now time.Time) error {
		stats, err := calcState(m.dur, now)
		if err != nil {
			return err
		}
		return event.Sent(stats)
	}, m.dur, true, false)

	return nil
}
