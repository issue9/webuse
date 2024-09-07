// SPDX-FileCopyrightText: 2015-2024 caixw
//
// SPDX-License-Identifier: MIT

//go:generate web locale -l=und -m -f=yaml ./
//go:generate web update-locale -src=./locales/und.yaml -dest=./locales/cmn-Hans.yaml

// Package webuse 适用 [web] 的中间件、插件和一些常用的路由函数
//
// [web]: https://github.com/issue9/web
package webuse
