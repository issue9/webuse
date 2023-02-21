// SPDX-License-Identifier: MIT

// Package sse server sent event 的实现
package sse

import "context"

const Mimetype = "text/event-stream"

// Server 事件管理
//
// T 表示用于区分不同事件源的 ID，比如按用户区分，那么该类型可能是 int64 类型的用户 ID 值。
type Server[T comparable] struct {
	status  int
	sources map[T]*source
}

func NewServer[T comparable](status int) *Server[T] {
	return &Server[T]{status: status}
}

func (e *Server[T]) Serve(ctx context.Context) error {
	e.sources = make(map[T]*source, 10)

	<-ctx.Done()
	for _, s := range e.sources {
		s.Close()
	}
	e.sources = nil
	return ctx.Err()
}
