// SPDX-License-Identifier: MIT

// Package switcher 按域名进行路由分类
package switcher

import "github.com/issue9/middleware/host"

// Switcher 域名切换中间件
//
// Deprecated: 已经不再使用，请使用 host.Switcher
type Switcher = host.Switcher

// New 声明新的 Switcher 实例
func New() *Switcher {
	return host.NewSwitcher()
}
