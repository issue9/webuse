// SPDX-License-Identifier: MIT

package health

import "net/http"

type response struct {
	http.ResponseWriter
	status int
}

func (r *response) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
