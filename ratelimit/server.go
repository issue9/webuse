// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package ratelimit

import (
	"errors"
	"net/http"
	"time"
)

// GenFunc 用于生成用户唯一 ID 的函数，用于区分令牌桶所属的用户。
type GenFunc func(*http.Request) (string, error)

// Store 存储 Bucket 的接口
type Store interface {
	// 设置或是添加指定名称的令牌桶
	Set(name string, val *Bucket) error

	// 删除指定的令牌桶
	Delete(name string) error

	// 获取指定的令牌桶，为空则返回 nil
	Get(name string) *Bucket
}

// Server 提供操作 Bucket 的一系列服务。
type Server struct {
	store    Store
	capacity int64
	rate     time.Duration
	genFunc  GenFunc
}

// GenIP 用于生成区分令牌桶的 IP 地址。
func GenIP(r *http.Request) (string, error) {
	if len(r.RemoteAddr) == 0 {
		return "", errors.New("无法获取请求端的 IP 地址")
	}
	return r.RemoteAddr, nil
}

// NewServer 声明一个新的 Server。
// fn 为令牌桶名称的产生方法，默认为用户的访问 IP。
func NewServer(store Store, capacity int64, rate time.Duration, fn GenFunc) *Server {
	if fn == nil {
		fn = GenIP
	}

	return &Server{
		store:    store,
		capacity: capacity,
		rate:     rate,
		genFunc:  fn,
	}
}

// 获取与前请求相对应的令牌桶。
func (srv *Server) bucket(r *http.Request) (*Bucket, error) {
	name, err := srv.genFunc(r)
	if err != nil {
		return nil, err
	}

	b := srv.store.Get(name)
	if b == nil {
		b = newBucket(srv.capacity, srv.rate)
		if err := srv.store.Set(name, b); err != nil {
			return nil, err
		}
	}

	return b, nil
}

// Transfer 将 oldName 的数据传送给 newName。
func (srv *Server) Transfer(oldName, newName string) error {
	b := srv.store.Get(oldName)
	if b != nil {
		if err := srv.store.Delete(oldName); err != nil {
			return err
		}
	}

	return srv.store.Set(newName, b)
}
