// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package rbac

import (
	"slices"

	"github.com/issue9/web"
)

// Role 角色信息
type Role[T comparable] struct {
	parent *Role[T]
	rbac   *RBAC[T]

	ID        string
	Parent    string
	Name      string
	Desc      string
	Resources []string // 当前角色关联的资源
	Users     []T      // 当前角色关联的用户
}

// 当前角色是否允许该用户 uid 访问资源 res
func (r *Role[T]) isAllow(uid T, res string) bool {
	return slices.Index(r.Users, uid) >= 0 && slices.Index(r.Resources, res) >= 0
}

// Add 添加角色信息
func (rbac *RBAC[T]) Add(name, desc string, parent string) (*Role[T], error) {
	role := &Role[T]{
		rbac:   rbac,
		parent: rbac.roles[parent],

		ID:        rbac.s.UniqueID(),
		Parent:    parent,
		Name:      name,
		Desc:      desc,
		Resources: make([]string, 0, 10),
		Users:     make([]T, 0, 10),
	}

	rbac.rolesMux.RLock()
	rbac.roles[role.ID] = role
	rbac.rolesMux.RUnlock()

	if err := rbac.store.Add(role); err != nil {
		return nil, err
	}
	return role, nil
}

// SetResources 关联角色与资源
//
// 替换之前关联的资源。如果传递空值，将直接清空 [Role.Resources]
func (role *Role[T]) SetResources(res ...string) error {
	if role.parent == nil {
		for _, resID := range res { // res 是否真实存在
			if !role.rbac.resourceExists(resID) {
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
	for _, child := range role.rbac.roles {
		if child.Parent != role.ID {
			continue
		}

		for _, childRes := range child.Resources {
			if slices.Index(res, childRes) < 0 {
				return web.NewLocaleError("child role has resource %s can not be deleted", childRes)
			}
		}
	}

	role.Resources = res
	return role.rbac.store.Set(role)
}

// Set 修改指定的角色信息
func (role *Role[T]) Set(name, desc string) error {
	if role.Name == name && role.Desc == role.Desc {
		return nil
	}

	role.Name = name
	role.Desc = desc
	return role.rbac.store.Set(role)
}

// Del 删除当前角色
func (role *Role[T]) Del() error {
	if len(role.Users) > 0 {
		return web.NewLocaleError("the role %s has users, can not deleted", role.ID)
	}

	role.rbac.rolesMux.RLock()
	for _, children := range role.rbac.roles {
		if children.Parent == role.ID {
			role.rbac.rolesMux.RUnlock()
			return web.NewLocaleError("the role %s has children role, can not deleted", role.ID)
		}
	}
	role.rbac.rolesMux.RUnlock()

	role.rbac.rolesMux.Lock()
	delete(role.rbac.roles, role.ID)
	role.rbac.rolesMux.Unlock()

	return role.rbac.store.Del(role.ID)
}

func (role *Role[T]) Link(uid T) error {
	if slices.Index(role.Users, uid) >= 0 { // 已经存在
		return nil
	}

	role.Users = append(role.Users, uid)
	return role.rbac.store.Set(role)
}

func (role *Role[T]) Unlink(uid T) error {
	if index := slices.Index(role.Users, uid); index >= 0 {
		role.Users = slices.Delete(role.Users, index, index+1)
		return role.rbac.store.Set(role)
	}

	return nil
}

// Roles 返回所有从当前角色继承的角色
//
// all 表示是否包含间接继承的角色
func (role *Role[T]) Roles(all bool) ([]*Role[T], error) {
	roles := make([]*Role[T], 0, 10)

	role.rbac.rolesMux.RLock()
	defer role.rbac.rolesMux.RUnlock()

	for _, r := range role.rbac.roles {
		if r.Parent == role.ID {
			roles = append(roles, r)
		}
	}

	if !all {
		return roles, nil
	}

	slices := roles

	for len(slices) > 0 { // [RBAC.Roles] 如果未传递值，则返回所有，所以得保证 slices 不为空
		rs, err := role.rbac.Roles(rolesID(slices)...)
		if err != nil {
			return nil, err
		}
		roles = append(roles, rs...)
		slices = rs
	}

	return roles, nil
}

func rolesID[T comparable](roles []*Role[T]) []string {
	keys := make([]string, 0, len(roles))
	for _, r := range roles {
		keys = append(keys, r.ID)
	}
	return keys
}

// Roles [Role.Parent] 在 p 中的角色列表
//
// 如果未指定 p，返回所有的角色列表。
func (rbac *RBAC[T]) Roles(p ...string) ([]*Role[T], error) {
	roles := make([]*Role[T], 0, len(rbac.roles))

	if len(p) == 0 {
		rbac.rolesMux.RLock()
		for _, v := range rbac.roles {
			roles = append(roles, v)
		}
		rbac.rolesMux.RUnlock()
	} else {
		rbac.rolesMux.RLock()
		for _, v := range rbac.roles {
			if slices.Index(p, v.Parent) >= 0 {
				roles = append(roles, v)
			}
		}
		rbac.rolesMux.RUnlock()
	}

	return roles, nil
}

// Role 返回指定的角色
//
// 如果找不到，则返回 nil
func (rbac *RBAC[T]) Role(id string) *Role[T] {
	rbac.rolesMux.RLock()
	defer rbac.rolesMux.RUnlock()
	return rbac.roles[id]
}
