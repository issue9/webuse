// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

//go:build !development

package debug

import "github.com/issue9/web"

// RegisterDev 仅在 [web.comptime.Mode] 为 development 时才注册
func RegisterDev(r *web.Router, path string) {}
