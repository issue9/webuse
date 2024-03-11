// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

// Package rbac RBAC 的简单实现
package rbac

import (
	"sync"

	"github.com/issue9/web"
)

// RBAC RBAC 实现
//
// T 表示的是用户的 ID 类型。
type RBAC[T comparable] struct {
	s         web.Server
	super     T // 超级管理员，不受权限组限制
	store     Store[T]
	info      *web.Logger
	getUID    GetUIDFunc[T]
	resources map[string]*Resources[T]
	roles     map[string]*Role[T]
	rolesMux  *sync.RWMutex
}

// GetUIDFunc 从 [web.Context] 获得当前的登录用户 ID
type GetUIDFunc[T comparable] func(*web.Context) (T, web.Responser)

// New 声明 [RBAC]
//
// super 表示超级管理员的 ID；
// info 用于输出一些提示信息，比如权限的判断依据等；
// getUID 参考 [GetUIDFunc]；
func New[T comparable](s web.Server, super T, store Store[T], info *web.Logger, getUID GetUIDFunc[T]) (*RBAC[T], error) {
	roles, err := store.Load()
	if err != nil {
		return nil, err
	}

	rbac := &RBAC[T]{
		s:         s,
		super:     super,
		store:     store,
		info:      info,
		getUID:    getUID,
		resources: make(map[string]*Resources[T], 50),
		rolesMux:  &sync.RWMutex{},
	}

	for _, role := range roles {
		role.rbac = rbac
	}
	rbac.roles = roles

	return rbac, nil
}

func (r *RBAC[T]) debug(uid T, res string, role *Role[T]) {
	r.info.LocaleString(web.Phrase("user %v obtained access to %s due to %s", uid, res, role.ID))
}
