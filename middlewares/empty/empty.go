// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

// Package empty 提供了一个不做任务附加操作的中间件
package empty

import "github.com/issue9/web"

// Empty 一个不做任何操作的中间件
//
// 在某些情况下需要保持中间件的变量是非空的值，可以采用此对象。
type Empty struct{}

func (m Empty) Middleware(next web.HandlerFunc, _, _, _ string) web.HandlerFunc { return next }
