// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package errorhandler

// 表示一个 HTTP 状态码错误。
// panic 此类型的值，可以在 Recovery 中作特殊处理。
//
// 目前仅由 Exit 使用，让框加以特定的状态码退出当前协程。
type httpStatus int

// Exit 退出当前的协程。
//
// status 表示输出的状态码，如果为 0，则不会作任何状态码输出。
//
// Exit 最终是以 panic 的形式退出，所以如果你的代码里截获了 panic，
// 那么 Exit 并不能达到退出当前请求的操作。
func Exit(status int) {
	panic(httpStatus(status))
}
