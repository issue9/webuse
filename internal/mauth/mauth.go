// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

// Package mauth middlewares/auth 的私有函数
package mauth

import "github.com/issue9/web"

type keyType int

const valueKey keyType = 1

// Set 更新 [web.Context] 保存的值
func Set[T any](ctx *web.Context, val T) { ctx.SetVar(valueKey, val) }

// Get 获取当前对话关联的信息
func Get[T any](ctx *web.Context) (val T, found bool) {
	if v, found := ctx.GetVar(valueKey); found {
		return v.(T), true
	}

	var zero T
	return zero, false
}
