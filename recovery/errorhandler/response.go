// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package errorhandler

import "net/http"

type response struct {
	http.ResponseWriter
}

func (r *response) WriteHeader(status int) {
	if status >= 400 {
		Exit(status)
	}

	r.ResponseWriter.WriteHeader(status)
}
