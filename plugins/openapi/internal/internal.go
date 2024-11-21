// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package internal

import (
	"encoding/json"
	"html/template"

	"gopkg.in/yaml.v3"
)

var Funcs = template.FuncMap{
	"json": func(v interface{}) template.JS {
		a, _ := json.Marshal(v)
		return template.JS(a)
	},

	"yaml": func(v interface{}) template.JS {
		a, _ := yaml.Marshal(v)
		return template.JS(a)
	},
}
