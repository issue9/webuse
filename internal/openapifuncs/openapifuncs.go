// SPDX-FileCopyrightText: 2024-2025 caixw
//
// SPDX-License-Identifier: MIT

package openapifuncs

import (
	"encoding/json"
	"html/template"

	"github.com/goccy/go-yaml"
)

var Funcs = template.FuncMap{
	"json": func(v any) template.JS {
		a, _ := json.Marshal(v)
		return template.JS(a)
	},

	"yaml": func(v any) template.JS {
		a, _ := yaml.Marshal(v)
		return template.JS(a)
	},
}
