// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package ratelimit

import (
	"net/http"
	"time"
)

// Store 存储 Bucket 的接口
type Store interface {
	Set(name string, val *Bucket) error
	Delete(name string) error
	Get(name string) (b *Bucket, found bool)
}

// Server 提供操作 Bucket 的一系列服务。
type Server struct {
	store    Store
	capacity int64
	rate     time.Duration
	getName  func(*http.Request) string
}

// NewServer 声明一个新的 Server
func NewServer(store Store, capacity int64, rate time.Duration, fn func(*http.Request) string) *Server {
	return &Server{
		store:    store,
		capacity: capacity,
		rate:     rate,
		getName:  fn,
	}
}

// Allow 当前请求是否被允许。
func (srv *Server) Allow(r *http.Request) (*Bucket, bool, error) {
	name := srv.getName(r)
	b, found := srv.store.Get(name)
	if !found {
		b = newBucket(srv.capacity, srv.rate)
	}

	allow := b.allow(1)

	if err := srv.store.Set(name, b); err != nil {
		return nil, false, err
	}

	return b, allow, nil
}

// Transfer 将 oldName 的数据传送给 newName。
func (srv *Server) Transfer(oldName, newName string) error {
	b, found := srv.store.Get(oldName)
	if found {
		if err := srv.store.Delete(oldName); err != nil {
			return err
		}
	}

	return srv.store.Set(newName, b)
}
