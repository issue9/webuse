// SPDX-License-Identifier: MIT

package middleware

import "net/http"

// Manager 中间件管理
type Manager struct {
	middlewares []Middleware

	// 下一步要执行的 Handler
	next http.Handler

	// 附加了中间件的 Handler
	handler http.Handler
}

// NewManager 声明新的 Manager 实例
func NewManager(next http.Handler) *Manager {
	return &Manager{
		middlewares: make([]Middleware, 0, 10),
		next:        next,
		handler:     next,
	}
}

// After 添加中间件到尾部
func (mgr *Manager) After(m Middleware) *Manager {
	mgr.middlewares = append(mgr.middlewares, m)
	mgr.handler = Handler(mgr.next, mgr.middlewares...)
	return mgr
}

// Before 添加中间件到顶部
func (mgr *Manager) Before(m Middleware) *Manager {
	ms := make([]Middleware, 0, 1+len(mgr.middlewares))
	ms = append(ms, m)
	if len(mgr.middlewares) > 0 {
		ms = append(ms, mgr.middlewares...)
	}
	mgr.middlewares = ms
	mgr.handler = Handler(mgr.next, mgr.middlewares...)
	return mgr
}

// Reset 清除中间件
func (mgr *Manager) Reset() *Manager {
	mgr.middlewares = mgr.middlewares[:0]
	mgr.handler = mgr.next
	return mgr
}

func (mgr *Manager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mgr.handler.ServeHTTP(w, r)
}
