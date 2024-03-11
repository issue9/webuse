// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package rbac

import (
	"fmt"
	"strings"

	"github.com/issue9/web"
)

// Resources 表示一组资源
type Resources[T comparable] struct {
	rbac      *RBAC[T]
	id        string
	title     web.LocaleStringer
	resources map[string]web.LocaleStringer
}

const idSeparator = '_'

// resourceExists 指定的资源 ID 是否存在
func (rbac *RBAC[T]) resourceExists(id string) bool {
	index := strings.IndexByte(id, idSeparator)
	if index < 0 {
		return false
	}

	gid := id[:index]
	if g, found := rbac.resources[gid]; found {
		_, f := g.resources[id]
		return f
	}
	return false
}

// NewResources 声明一组资源
//
// id 为该资源组的唯一 ID；
// title 对该资源组的描述；
func (rbac *RBAC[T]) NewResources(id string, title web.LocaleStringer) *Resources[T] {
	if _, found := rbac.resources[id]; found {
		panic(fmt.Sprintf("已经存在同名的资源组 %s", id))
	}

	res := &Resources[T]{
		rbac:      rbac,
		id:        id,
		title:     title,
		resources: make(map[string]web.LocaleStringer, 10),
	}
	rbac.resources[id] = res
	return res
}

// New 添加新的资源
//
// 返回的是用于判断是否拥有当前资源权限的中间件。
func (r *Resources[T]) New(id string, desc web.LocaleStringer) web.MiddlewareFunc {
	id = r.id + string(idSeparator) + id

	if _, found := r.resources[id]; found {
		panic(fmt.Sprintf("已经存在同名的资源 %s", id))
	}
	r.resources[id] = desc

	return func(next web.HandlerFunc) web.HandlerFunc {
		return func(ctx *web.Context) web.Responser {
			uid, resp := r.rbac.getUID(ctx)
			if resp != nil {
				return resp
			}

			if uid == r.rbac.super {
				return next(ctx)
			}

			for _, role := range r.rbac.roles {
				if role.isAllow(uid, id) {
					r.rbac.debug(uid, id, role)
					return next(ctx)
				}
			}

			return ctx.Problem(web.ProblemForbidden)
		}
	}
}
