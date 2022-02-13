// SPDX-License-Identifier: MIT

package health

import (
	"errors"
	"log"
	"sort"

	"github.com/issue9/cache"
	"github.com/issue9/sliceutil"
)

type cacheStore struct {
	errlog *log.Logger
	cache  cache.Cache
	prefix string
	allKey string
}

// NewCache 将 cache 作为存储介质
//
// prefix 存储时，统一的前缀名称，防止重名；
// errlog 当存储出错时，错误信息将保存到 errlog；
func NewCache(c cache.Cache, prefix string, errlog *log.Logger) Store {
	allKey := prefix + "_all_key"
	if err := c.Set(allKey, []string{}, cache.Forever); err != nil {
		errlog.Println(err)
	}

	return &cacheStore{
		errlog: errlog,
		cache:  c,
		prefix: prefix,
		allKey: allKey,
	}
}

func (c *cacheStore) getID(method, path string) string {
	return c.prefix + method + "_" + path
}

func (c *cacheStore) Get(method, path string) *State {
	key := c.getID(method, path)

	s, err := c.cache.Get(key)
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
	if err := c.cache.Set(key, state, cache.Forever); err != nil {
		c.errlog.Println(err)
	}

	all := c.keys()
	if sliceutil.Index(all, func(e string) bool { return e == key }) > -1 {
		return
	}
	all = append(all, key)
	sort.Strings(all)
	if err := c.cache.Set(c.allKey, all, cache.Forever); err != nil {
		c.errlog.Println(err)
	}
}

func (c *cacheStore) keys() []string {
	allInterface, err := c.cache.Get(c.allKey)
	if err != nil {
		c.errlog.Println(err)
		return nil
	}
	return allInterface.([]string)
}

func (c *cacheStore) All() []*State {
	all := c.keys()
	states := make([]*State, 0, len(all))
	for _, key := range all {
		s, err := c.cache.Get(key)
		if err != nil {
			c.errlog.Println(err)
			continue
		}
		states = append(states, s.(*State))
	}
	return states
}
