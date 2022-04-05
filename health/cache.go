// SPDX-License-Identifier: MIT

package health

import (
	"errors"
	"sort"

	"github.com/issue9/cache"
	"github.com/issue9/sliceutil"
	"github.com/issue9/web"
)

const allKey = "all_key"

type cacheStore struct {
	access cache.Access
	errlog web.Logger
}

func newCache(srv *web.Server, prefix string) Store {
	access := cache.Prefix(prefix, srv.Cache())
	errlog := srv.Logs().ERROR()
	if err := access.Set(allKey, []string{}, cache.Forever); err != nil {
		errlog.Error(err)
	}

	return &cacheStore{
		access: access,
		errlog: errlog,
	}
}

func (c *cacheStore) getID(method, path string) string { return method + "_" + path }

func (c *cacheStore) Get(method, path string) *State {
	key := c.getID(method, path)

	s, err := c.access.Get(key)
	if errors.Is(err, cache.ErrCacheMiss) {
		state := &State{
			Method: method,
			Path:   path,
		}
		c.Save(state)
		return state
	}

	return s.(*State)
}

func (c *cacheStore) Save(state *State) {
	key := c.getID(state.Method, state.Path)
	if err := c.access.Set(key, state, cache.Forever); err != nil {
		c.errlog.Error(err)
	}

	all := c.keys()
	if sliceutil.Index(all, func(e string) bool { return e == key }) > -1 {
		return
	}
	all = append(all, key)
	sort.Strings(all)
	if err := c.access.Set(allKey, all, cache.Forever); err != nil {
		c.errlog.Error(err)
	}
}

func (c *cacheStore) keys() []string {
	allInterface, err := c.access.Get(allKey)
	if err != nil {
		c.errlog.Error(err)
		return nil
	}
	return allInterface.([]string)
}

func (c *cacheStore) All() []*State {
	all := c.keys()
	states := make([]*State, 0, len(all))
	for _, key := range all {
		s, err := c.access.Get(key)
		if err != nil {
			c.errlog.Error(err)
			continue
		}
		states = append(states, s.(*State))
	}
	return states
}
