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
	sources map[T]*Source
}

func NewServer[T comparable](status int) *Server[T] {
	return &Server[T]{status: status}
}

func (srv *Server[T]) Serve(ctx context.Context) error {
	srv.sources = make(map[T]*Source, 10)

	<-ctx.Done()
	for _, s := range srv.sources {
		s.Close()
	}
	srv.sources = nil
	return ctx.Err()
}

// Len 当前活动的数量
func (srv *Server[T]) Len() int { return len(srv.sources) }
