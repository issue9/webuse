// SPDX-License-Identifier: MIT

package errorhandler

import "net/http"

// 表示一个 HTTP 状态码错误
//
// panic 此类型的值，可以在 Recovery 中作特殊处理。
//
// 目前仅由 Exit 使用，让框加以特定的状态码退出当前协程。
type httpStatus int

type response struct {
	http.ResponseWriter
	eh *ErrorHandler
}

func (r *response) WriteHeader(status int) {
	if r.eh.Exists(status) {
		r.eh.Exit(r.ResponseWriter, status)
		return
	}
	r.ResponseWriter.WriteHeader(status)
}

// WriteHeader 写入 HTTP 状态值
//
// 通过 errorhandler 的封装之后，默认会将注册的状态码 响应重定向到指定的处理函数，
// 如果不需要特殊处理， 可以调用此函数，按照正常流程处理。
func WriteHeader(w http.ResponseWriter, status int) {
	if resp, ok := w.(*response); ok {
		resp.ResponseWriter.WriteHeader(status)
		return
	}
	w.WriteHeader(status)
}

// Exit 以指定的状态码退出当前的协程
//
// Exit 最终是以 panic 的形式退出，所以如果你的代码里截获了 panic，
// 那么 Exit 并不能达到退出当前请求的操作。
func (e *ErrorHandler) Exit(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
	e.Render(w, status)
	panic(httpStatus(0))
}
