// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

// Package rbac RBAC 的简单实现
//
//	rbac := rbac.New(...)
//	group := rbac.NewGroup("user", web.Phrase("user"))
//
//	view := group.New("view", web.Phrase("view info")) // 返回判断权限的中间件
//	del := group.New("del", web.Phrase("delete user")) // 返回判断权限的中间件
//	router.Get("/users", view(func(*web.Context)web.Responser{
//	    // do somthing
//	}))
//
//	router.Delete("/users/{id}", del(func(*web.Context)web.Responser{
//	    // do somthing
//	}))
package rbac

import (
	"sync"

	"github.com/issue9/web"
)

// RBAC RBAC 实现
//
// T 表示的是用户的 ID 类型。
type RBAC[T comparable] struct {
	s      web.Server
	super  T // 超级管理员，不受权限组限制
	store  Store[T]
	info   *web.Logger
	getUID GetUIDFunc[T]

	resources []string // 缓存的所有资源项
	groups    map[string]*Group[T]

	roles    map[string]*Role[T]
	rolesMux *sync.RWMutex
}

// GetUIDFunc 从 [web.Context] 获得当前的登录用户 ID
type GetUIDFunc[T comparable] func(*web.Context) (T, web.Responser)

// New 声明 [RBAC]
//
// super 表示超级管理员的 ID；
// info 用于输出一些提示信息，比如权限的判断依据等；
// getUID 参考 [GetUIDFunc]；
// loadInterval 如果大于 0，表示以该频率从 [Store] 加载数据；
func New[T comparable](s web.Server, super T, store Store[T], info *web.Logger, getUID GetUIDFunc[T]) (*RBAC[T], error) {
	rbac := &RBAC[T]{
		s:      s,
		super:  super,
		store:  store,
		info:   info,
		getUID: getUID,

		resources: make([]string, 0, 100),
		groups:    make(map[string]*Group[T], 50),

		rolesMux: &sync.RWMutex{},
	}

	if err := rbac.Load(); err != nil {
		return nil, err
	}

	return rbac, nil
}

// Load 加载数据
func (rbac *RBAC[T]) Load() error {
	roles, err := rbac.store.Load()
	if err != nil {
		return err
	}

	for _, role := range roles {
		role.rbac = rbac
	}

	rbac.rolesMux.Lock()
	rbac.roles = roles
	rbac.rolesMux.Unlock()

	return nil
}

func (r *RBAC[T]) debug(uid T, res string, role *Role[T]) {
	r.info.LocaleString(web.Phrase("user %v obtained access to %s due to %s", uid, res, role.ID))
}
