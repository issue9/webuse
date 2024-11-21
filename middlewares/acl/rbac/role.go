// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package rbac

import (
	"fmt"
	"iter"
	"slices"
	"sync"

	"github.com/issue9/web"
)

// RoleGroup 角色分组
//
// 当一个用户系统有多个独立的权限模块时，可以很好地用 [RoleGroup] 进行表示，
// 比如商家系统，每个商家拥有自己的操作人员，可为每个商家创建 [RoleGroup]。
type RoleGroup[T comparable] struct {
	rbac    *RBAC[T]
	id      string
	superID T

	roles    map[string]*Role[T]
	rolesMux *sync.RWMutex
}

// Role 角色信息
type Role[T comparable] struct {
	parent *Role[T]
	group  *RoleGroup[T]

	ID        string
	Parent    string
	Name      string
	Desc      string
	Resources []string // 当前角色关联的资源
	Users     []T      // 当前角色关联的用户
}

// NewRoleGroup 声明 [RoleGroup]
//
// id 表示当前角色组的唯一 ID；
// superID 表示超级管理员的 ID；
func (rbac *RBAC[T]) NewRoleGroup(id string, superID T) (*RoleGroup[T], error) {
	if _, found := rbac.roleGroups[id]; found {
		panic(fmt.Sprintf("%s 已经存在", id))
	}

	g := &RoleGroup[T]{
		rbac:    rbac,
		id:      id,
		superID: superID,

		roles:    make(map[string]*Role[T], 10),
		rolesMux: &sync.RWMutex{},
	}
	if err := g.Load(); err != nil {
		return nil, err
	}
	rbac.roleGroups[id] = g

	return g, nil
}

// UserRoles 用户 uid 关联的角色列表
func (g *RoleGroup[T]) UserRoles(uid T) []*Role[T] {
	roles := make([]*Role[T], 0, 5)
	for _, r := range g.roles {
		if slices.Index(r.Users, uid) >= 0 {
			roles = append(roles, r)
		}
	}
	return roles
}

func (g *RoleGroup[T]) RBAC() *RBAC[T] { return g.rbac }

// Load 加载数据
func (g *RoleGroup[T]) Load() error {
	roles, err := g.RBAC().store.Load(g.id)
	if err != nil {
		return err
	}

	for _, role := range roles {
		role.group = g
	}

	g.rolesMux.Lock()
	g.roles = roles
	g.rolesMux.Unlock()

	return nil
}

// NewRole 添加角色信息
func (g *RoleGroup[T]) NewRole(name, desc, parent string) (*Role[T], error) {
	role := &Role[T]{
		group:  g,
		parent: g.Role(parent),

		ID:        g.rbac.s.UniqueID(),
		Parent:    parent,
		Name:      name,
		Desc:      desc,
		Resources: make([]string, 0, 10),
		Users:     make([]T, 0, 10),
	}

	g.rolesMux.RLock()
	g.roles[role.ID] = role
	g.rolesMux.RUnlock()

	if err := g.rbac.store.Add(g.id, role); err != nil {
		return nil, err
	}
	return role, nil
}

// Roles 当前的所有角色
func (g *RoleGroup[T]) Roles() iter.Seq[*Role[T]] {
	return func(yield func(*Role[T]) bool) {
		g.rolesMux.RLock()
		for _, v := range g.roles {
			if !yield(v) {
				break
			}
		}
		g.rolesMux.RUnlock()
	}
}

// Role 返回指定的角色
//
// 如果找不到，则返回 nil
func (g *RoleGroup[T]) Role(id string) *Role[T] {
	g.rolesMux.RLock()
	r := g.roles[id]
	g.rolesMux.RUnlock()
	return r
}

// 用户 uid 是否可访问资源 res
func (g *RoleGroup[T]) isAllow(uid T, res string) bool {
	if uid == g.superID {
		return true
	}

	g.rolesMux.RLock()
	defer g.rolesMux.RUnlock()

	for _, role := range g.roles {
		if slices.Index(role.Users, uid) >= 0 && slices.Index(role.Resources, res) >= 0 {
			msg := web.Phrase("user %v obtained access to %s due to %s:%s", uid, res, role.ID, g.id)
			g.rbac.s.Logs().INFO().LocaleString(msg)
			return true
		}
	}

	return false
}

// Allow 关联角色与资源
//
// 替换之前关联的资源。如果传递空值，将直接清空 [Role.Resources]
func (role *Role[T]) Allow(res ...string) error {
	if role.parent == nil {
		for _, resID := range res { // res 是否真实存在
			if !role.group.rbac.resourceExists(resID) {
				return web.NewLocaleError("not found resource %s", resID)
			}
		}
	} else {
		for _, resID := range res {
			if slices.Index(role.parent.Resources, resID) < 0 {
				return web.NewLocaleError("not found resource %s", resID)
			}
		}
	}

	// 判断子角色中的资源是否都在 res 之中
	role.group.rolesMux.RLock()
	for _, child := range role.group.roles {
		if child.Parent != role.ID {
			continue
		}

		for _, childRes := range child.Resources {
			if slices.Index(res, childRes) < 0 {
				return web.NewLocaleError("child role has resource %s can not be deleted", childRes)
			}
		}
	}
	role.group.rolesMux.RUnlock()

	role.Resources = res
	return role.group.rbac.store.Set(role.group.id, role)
}

// Set 修改指定的角色信息
func (role *Role[T]) Set(name, desc string) error {
	if role.Name == name && role.Desc == desc {
		return nil
	}

	role.Name = name
	role.Desc = desc
	return role.group.rbac.store.Set(role.group.id, role)
}

// Del 删除当前角色
func (role *Role[T]) Del() error {
	if len(role.Users) > 0 {
		return web.NewLocaleError("the role %s has users, can not deleted", role.ID)
	}

	role.group.rolesMux.RLock()
	for _, children := range role.group.roles {
		if children.Parent == role.ID {
			role.group.rolesMux.RUnlock()
			return web.NewLocaleError("the role %s has children role, can not deleted", role.ID)
		}
	}
	role.group.rolesMux.RUnlock()

	role.group.rolesMux.Lock()
	delete(role.group.roles, role.ID)
	role.group.rolesMux.Unlock()

	return role.group.rbac.store.Del(role.group.id, role.ID)
}

func (role *Role[T]) Link(uid T) error {
	if slices.Index(role.Users, uid) >= 0 { // 已经存在于当前用户
		return nil
	}

	parent := role.parent
	for parent != nil {
		if slices.Index(parent.Users, uid) >= 0 { // 已经存在于父角色
			return web.NewLocaleError("user %v in the parent role %s", uid, parent.ID)
		}
		parent = parent.parent
	}

	role.Users = append(role.Users, uid)
	return role.group.rbac.store.Set(role.group.id, role)
}

func (role *Role[T]) Unlink(uid T) error {
	if index := slices.Index(role.Users, uid); index >= 0 {
		role.Users = slices.Delete(role.Users, index, index+1)
		return role.group.rbac.store.Set(role.group.id, role)
	}

	return nil
}

// IsDescendant 判断角色 rid 是否为当前角色的子角色
func (role *Role[T]) IsDescendant(rid string) bool {
	role.group.rolesMux.RLock()
	defer role.group.rolesMux.RUnlock()

	for _, r := range role.group.roles {
		if r.Parent != role.ID {
			continue
		}

		if r.ID == rid || r.IsDescendant(rid) {
			return true
		}
	}

	return false
}

// Descendants 返回所有从当前角色继承的角色
//
// all 表示是否包含间接继承的角色
func (role *Role[T]) Descendants(all bool) []*Role[T] {
	roles := make([]*Role[T], 0, 10)

	role.group.rolesMux.RLock()
	defer role.group.rolesMux.RUnlock()

	for _, r := range role.group.roles {
		if r.Parent == role.ID {
			roles = append(roles, r)
		}
	}

	if !all {
		return roles
	}

	s := roles
	for len(s) > 0 {
		ids := rolesID(s)
		rs := make([]*Role[T], 0, 10)

		for _, v := range role.group.roles {
			if slices.Index(ids, v.Parent) >= 0 {
				rs = append(rs, v)
			}
		}
		roles = append(roles, rs...)
		s = rs
	}

	return roles
}

func rolesID[T comparable](roles []*Role[T]) []string {
	keys := make([]string, 0, len(roles))
	for _, r := range roles {
		keys = append(keys, r.ID)
	}
	return keys
}
