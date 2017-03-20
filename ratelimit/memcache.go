// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package ratelimit

import (
	"encoding/json"
	"log"

	gm "github.com/bradfitz/gomemcache/memcache"
)

type memcache struct {
	prefix     string
	client     *gm.Client
	errlog     *log.Logger
	expiration int32
}

// NewMemcache 声明一个以 memcache 为存储介质的 Store
//
// prefix 变量名的统一前缀，防止与其它变量名冲突
// client memcached 客户端实例。
// expiration 过期时间，单位：秒。
// errlog 错误时，写的日志记录。
func NewMemcache(prefix string, client *gm.Client, expiration int32, errlog *log.Logger) Store {
	return &memcache{
		prefix:     prefix,
		client:     client,
		expiration: expiration,
		errlog:     errlog,
	}
}

func (mem *memcache) Set(name string, b *Bucket) error {
	data, err := json.Marshal(b)
	if err != nil {
		return err
	}

	return mem.client.Set(&gm.Item{
		Key:        mem.prefix + name,
		Value:      data,
		Expiration: mem.expiration,
	})
}

func (mem *memcache) Delete(name string) error {
	return mem.client.Delete(mem.prefix + name)
}

func (mem *memcache) Get(name string) (*Bucket, bool) {
	item, err := mem.client.Get(mem.prefix + name)
	if err != nil {
		if mem.errlog != nil {
			mem.errlog.Println(err)
		}
		return nil, false
	}

	b := &Bucket{}
	if err = json.Unmarshal(item.Value, b); err != nil {
		mem.errlog.Println(err)
		return nil, false
	}

	return b, true
}
