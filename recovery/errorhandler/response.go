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

// WriteHeader 如果不想让 400 以上的状态码被作特殊处理，
// 可以调用此方法逃避。
func WriteHeader(w http.ResponseWriter, status int) {
	if status < 400 {
		w.WriteHeader(status)
		return
	}

	if resp, ok := w.(*response); ok {
		resp.ResponseWriter.WriteHeader(status)
		return
	}

	w.WriteHeader(status)
}
