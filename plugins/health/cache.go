// SPDX-FileCopyrightText: 2015-2024 caixw
//
// SPDX-License-Identifier: MIT

package health

import (
	"errors"
	"slices"

	"github.com/issue9/cache"
	"github.com/issue9/web"
)

type cacheStore struct {
	cache  web.Cache
	errlog *web.Logger
}

// NewCacheStore 基于缓存的存取接口实现
//
// NOTE: 缓存是易失性的，不能永久性保存数据。
func NewCacheStore(srv web.Server, prefix string) Store {
	c := web.NewCache(prefix, srv.Cache())
	errlog := srv.Logs().ERROR()
	if err := c.Set("", []string{}, cache.Forever); err != nil {
		errlog.Error(err)
	}

	return &cacheStore{
		cache:  c,
		errlog: errlog,
	}
}

func (c *cacheStore) getID(route, method, pattern string) string {
	return route + "_" + method + "_" + pattern
}

func (c *cacheStore) Get(route, method, pattern string) *State {
	key := c.getID(route, method, pattern)

	s := &State{}
	if err := c.cache.Get(key, s); errors.Is(err, cache.ErrCacheMiss()) {
		state := newState(route, method, pattern)
		c.Save(state)
		return state
	}

	return s
}

func (c *cacheStore) Save(state *State) {
	key := c.getID(state.Route, state.Method, state.Pattern)
	if err := c.cache.Set(key, state, cache.Forever); err != nil {
		c.errlog.Error(err)
	}

	all := c.keys()
	if slices.Index(all, key) >= 0 {
		return
	}
	all = append(all, key)
	slices.Sort(all)
	if err := c.cache.Set("", all, cache.Forever); err != nil {
		c.errlog.Error(err)
	}
}

func (c *cacheStore) keys() []string {
	all := []string{}
	if err := c.cache.Get("", &all); err != nil {
		c.errlog.Error(err)
		return nil
	}
	return all
}

func (c *cacheStore) All() []*State {
	all := c.keys()
	states := make([]*State, 0, len(all))
	for _, key := range all {
		s := &State{}
		if err := c.cache.Get(key, s); err != nil {
			c.errlog.Error(err)
			continue
		}
		states = append(states, s)
	}
	return states
}
