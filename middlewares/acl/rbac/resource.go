// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package rbac

import (
	"cmp"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/issue9/web"
	"golang.org/x/text/message"
)

const idSeparator = '_'

// ResourceGroup 资源组
type ResourceGroup[T comparable] struct {
	rbac  *RBAC[T]
	id    string
	title web.LocaleStringer
	items map[string]web.LocaleStringer
}

// Resources 资源
type Resources struct {
	ID    string      `json:"id" xml:"id,attr" yaml:"id" cbor:"id" comment:"the id of resource group"`
	Title string      `json:"title" xml:"title" yaml:"title" cbor:"title" comment:"the title of resource group"`
	Items []*Resource `json:"items,omitempty" xml:"items>item,omitempty" yaml:"items,omitempty" cbor:"items,omitempty" comment:"resources"`
}

// Resource 单个资源
type Resource struct {
	ID    string `json:"id" xml:"id,attr" yaml:"id" cbor:"id" comment:"the id of resource"`
	Title string `json:"title" xml:"title" yaml:"title" cbor:"title" comment:"the title of resource"`
}

// RoleResources 表示某个角色所能访问的资源
type RoleResources struct {
	// Current 角色当前能访问的资源
	Current []string `json:"current" xml:"current" yaml:"current" cbor:"current" comment:"resources of current role"`

	// Parent 角色的父类能访问的资源
	//
	// 必然也是当前角色所能访问的最大资源列表。
	//
	// Parent 肯定是包含了 Current 的所有值。
	Parent []string `json:"parent" xml:"parent" yaml:"parent" cbor:"parent" comment:"resources of role parent"`
}

// resourceExists 指定的资源 ID 是否存在
func (rbac *RBAC[T]) resourceExists(id string) bool {
	index := strings.IndexByte(id, idSeparator)
	if index < 0 {
		return false
	}

	gid := id[:index]
	if g, found := rbac.resourceGroups[gid]; found {
		_, f := g.items[id]
		return f
	}
	return false
}

// NewResourceGroup 声明一组资源
//
// id 为该资源组的唯一 ID；
// title 对该资源组的描述；
func (rbac *RBAC[T]) NewResourceGroup(id string, title web.LocaleStringer) *ResourceGroup[T] {
	if _, found := rbac.resourceGroups[id]; found {
		panic(fmt.Sprintf("已经存在同名的资源组 %s", id))
	}

	res := &ResourceGroup[T]{
		rbac:  rbac,
		id:    id,
		title: title,
		items: make(map[string]web.LocaleStringer, 10),
	}
	rbac.resourceGroups[id] = res
	return res
}

func (rbac *RBAC[T]) ResourceGroup(id string) *ResourceGroup[T] { return rbac.resourceGroups[id] }

func joinID(gid, id string) string { return gid + string(idSeparator) + id }

func (g *ResourceGroup[T]) RBAC() *RBAC[T] { return g.rbac }

// New 添加新的资源
//
// 返回的是用于判断是否拥有当前资源权限的中间件。
func (g *ResourceGroup[T]) New(id string, desc web.LocaleStringer) web.MiddlewareFunc {
	id = joinID(g.id, id)

	if _, found := g.items[id]; found {
		panic(fmt.Sprintf("已经存在同名的资源 %s", id))
	}
	g.items[id] = desc
	g.RBAC().resources = append(g.rbac.resources, id)

	return func(next web.HandlerFunc, method, _, _ string) web.HandlerFunc {
		if method == http.MethodOptions {
			return next
		}

		return func(ctx *web.Context) web.Responser {
			uid, resp := g.rbac.getUID(ctx)
			if resp != nil {
				return resp
			}

			for _, roleG := range g.rbac.roleGroups {
				if roleG.isAllow(uid, id) {
					return next(ctx)
				}
			}

			return ctx.Problem(web.ProblemForbidden)
		}
	}
}

// Resources 所有资源的列表
func (rbac *RBAC[T]) Resources(p *message.Printer) []*Resources {
	res := make([]*Resources, 0, len(rbac.resourceGroups))
	for _, role := range rbac.resourceGroups {
		items := make([]*Resource, 0, len(role.items))
		for id, item := range role.items {
			items = append(items, &Resource{ID: id, Title: item.LocaleString(p)})
		}
		slices.SortFunc(items, func(a, b *Resource) int { return cmp.Compare(a.ID, b.ID) })

		res = append(res, &Resources{
			ID:    role.id,
			Title: role.title.LocaleString(p),
			Items: items,
		})
		slices.SortFunc(res, func(a, b *Resources) int { return cmp.Compare(a.ID, b.ID) })
	}
	return res
}

// Resource 当前角色的资源信息
func (role *Role[T]) Resource() *RoleResources {
	var parent []string
	if role.parent == nil {
		parent = role.group.rbac.resources
	} else {
		parent = role.parent.Resources
	}

	var current []string
	if len(role.Resources) > 0 {
		current = role.Resources
	}

	return &RoleResources{
		Current: slices.Clone(current),
		Parent:  slices.Clone(parent),
	}
}
