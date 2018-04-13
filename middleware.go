// Copyright 2015 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package middleware 包含了一系列 http.Handler 接口的中间件。
package middleware

import "net/http"

// Middleware 将一个 http.Handler 封装成另一个 http.Handler
type Middleware func(http.Handler) http.Handler
