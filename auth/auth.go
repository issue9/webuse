// SPDX-License-Identifier: MIT

// Package auth 验证类的中间件
package auth

import "github.com/issue9/web"

type keyType int

const valueKey keyType = 0

func SetValue(ctx *web.Context, v interface{}) { ctx.Vars[valueKey] = v }

func GetValue(ctx *web.Context) (v interface{}, found bool) {
	v, found = ctx.Vars[valueKey]
	return
}
