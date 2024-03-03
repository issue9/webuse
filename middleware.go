// SPDX-FileCopyrightText: 2015-2024 caixw
//
// SPDX-License-Identifier: MIT

//go:generate web locale -l=und -m -f=yaml ./
//go:generate web update-locale -src=./locales/und.yaml -dest=./locales/zh-Hans.yaml

// Package middleware 适用 [web] 的一些中间件实现
//
// [web]: https://github.com/issue9/web
package middleware
