// SPDX-License-Identifier: MIT

package errorhandler

import "net/http"

type response struct {
	http.ResponseWriter
	eh *ErrorHandler
}

func (r *response) WriteHeader(status int) {
	if _, found := r.eh.handlers[status]; found { // 已经定义该状态码相应的处理方式
		Exit(status)
	}

	r.ResponseWriter.WriteHeader(status)
}

// WriteHeader 写入 HTTP 状态值
//
// 通过 errorhandler 的封装之后，默认会将大于等于 400
// 状态值的响应重定向到指定的处理函数，如果不需要特殊处理，
// 可以调用此函数，按照正常流程处理。
func WriteHeader(w http.ResponseWriter, status int) {
	if resp, ok := w.(*response); ok {
		resp.ResponseWriter.WriteHeader(status)
		return
	}
	w.WriteHeader(status)
}
