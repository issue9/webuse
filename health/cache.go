// SPDX-License-Identifier: MIT

package health

import (
	"errors"
	"sort"

	"github.com/issue9/sliceutil"
	"github.com/issue9/web"
	"github.com/issue9/web/cache"
)

const allKey = "all_key"

type cacheStore struct {
	cache  web.Cache
	errlog web.Logger
}

// NewCacheStore 基于缓存的存取接口实现
//
// NOTE: 缓存是易失性的，不能永久性保存数据。
func NewCacheStore(srv *web.Server, prefix string) Store {
	access := cache.Prefix(srv.Cache(), prefix+"_")
	errlog := srv.Logs().ERROR()
	if err := access.Set(allKey, []string{}, cache.Forever); err != nil {
		errlog.Error(err)
	}

	return &cacheStore{
		cache:  access,
		errlog: errlog,
	}
}

func (c *cacheStore) getID(method, path string) string { return method + "_" + path }

func (c *cacheStore) Get(method, pattern string) *State {
	key := c.getID(method, pattern)

	s := &State{}
	err := c.cache.Get(key, s)
	if errors.Is(err, cache.ErrCacheMiss()) {
		state := newState(method, pattern)
		c.Save(state)
		return state
	}

	return s
}

func (c *cacheStore) Save(state *State) {
	key := c.getID(state.Method, state.Pattern)
	if err := c.cache.Set(key, state, cache.Forever); err != nil {
		c.errlog.Error(err)
	}

	all := c.keys()
	if sliceutil.Index(all, func(e string) bool { return e == key }) > -1 {
		return
	}
	all = append(all, key)
	sort.Strings(all)
	if err := c.cache.Set(allKey, all, cache.Forever); err != nil {
		c.errlog.Error(err)
	}
}

func (c *cacheStore) keys() []string {
	allInterface := []string{}
	if err := c.cache.Get(allKey, &allInterface); err != nil {
		c.errlog.Error(err)
		return nil
	}
	return allInterface
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
