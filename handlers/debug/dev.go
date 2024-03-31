// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

//go:build development

package debug

import "github.com/issue9/web"

func Init(r *web.Router, path, name string) {
	r.Get(path, New(name, web.ProblemNotFound))
}
