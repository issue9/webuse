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

import "github.com/issue9/web"

// RBAC RBAC 实现
//
// T 表示的是用户的 ID 类型。
type RBAC[T comparable] struct {
	s      web.Server
	store  Store[T]
	getUID GetUIDFunc[T]

	resources      []string // 缓存的所有资源项
	resourceGroups map[string]*ResourceGroup[T]

	roleGroups map[string]*RoleGroup[T]
}

// GetUIDFunc 从 [web.Context] 获得当前的登录用户 ID
type GetUIDFunc[T comparable] func(*web.Context) (T, web.Responser)

// New 声明 [RBAC]
//
// getUID 参考 [GetUIDFunc]；
// loadInterval 如果大于 0，表示以该频率从 [Store] 加载数据；
func New[T comparable](s web.Server, store Store[T], getUID GetUIDFunc[T]) *RBAC[T] {
	return &RBAC[T]{
		s:      s,
		store:  store,
		getUID: getUID,

		resources:      make([]string, 0, 100),
		resourceGroups: make(map[string]*ResourceGroup[T], 50),

		roleGroups: make(map[string]*RoleGroup[T], 50),
	}
}
