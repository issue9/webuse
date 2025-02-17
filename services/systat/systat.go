// SPDX-FileCopyrightText: 2024-2025 caixw
//
// SPDX-License-Identifier: MIT

// Package systat 系统状态检测
package systat

import (
	"container/ring"
	"context"
	"time"

	"github.com/issue9/events"
	"github.com/issue9/web"
)

// 系统状态监视服务
type service struct {
	events *events.Event[*Stats]
	ring   *ring.Ring
}

// Init 初始化监视系统状态的服务
//
// 这将返回一个用于订阅状态变化的接口，用户可根据该接口订阅信息。
//
// dur 为监视数据的频率；
// interval 为每次监视数据的时间，可以为 0，表示从最后一次调用开始计算；
// size 缓存监控数据的数量；
func Init(s web.Server, dur, interval time.Duration, size int) events.Subscriber[*Stats] {
	srv := &service{
		events: events.New[*Stats](),
		ring:   ring.New(size),
	}

	job := func(now time.Time) error {
		if srv.events.Len() == 0 { // 没有订阅，就没必要计算状态了。
			return nil
		}

		stat, err := calcState(interval, now)
		if err == nil {
			srv.ring.Value = stat
			srv.ring = srv.ring.Next()
			srv.events.Publish(true, stat)
		}
		return err
	}

	// 此时 events.Len 必然为空，没必要将 true 传递给 imm。
	s.Services().AddTicker(web.Phrase("monitor system stat"), job, dur, false, true)

	return srv
}

// Subscribe 订阅状态变化的通知
func (s *service) Subscribe(f events.SubscribeFunc[*Stats]) context.CancelFunc {
	s.ring.Next().Do(func(v any) { // 一次性发送所有缓存的数据
		if v != nil {
			f(v.(*Stats))
		}
	})

	return s.events.Subscribe(f)
}
