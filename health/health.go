// SPDX-License-Identifier: MIT

// Package health API 状态检测
package health

import (
	"net/http"
	"time"
)

// Store 存储 API 状态的处理接口
type Store interface {
	// 获取指定 API 的数据
	//
	// 如果还不存在，则返回空对象。
	Get(method, path string) *State

	// 每生成一条数据，均会以异步的方式调用 Save，由处理具体的操作方式。
	Save(*State)

	// 返回所有接口的状态信息
	All() []*State
}

// State 实际存在的数据类型
type State struct {
	Method, Path string
	Min, Max     time.Duration
	Count        int           // 总的请求次数
	UserErrors   int           // 用户端出错次数，400-499
	ServerErrors int           // 服务端出错次数，>500
	Last         time.Time     // 最后的访问时间
	Spend        time.Duration // 总花费的时间
}

// Health API 状态检测
type Health struct {
	Enabled bool // 是否启用当前的中间件
	store   Store
}

// New 声明 Health 实例
func New(store Store) *Health {
	return &Health{
		Enabled: true,
		store:   store,
	}
}

// States 返回所有的状态列表
func (h *Health) States() []*State {
	return h.store.All()
}

// MiddlewareFunc 将当前中间件应用于 next
func (h *Health) MiddlewareFunc(next func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return h.Middleware(http.HandlerFunc(next))
}

// Middleware 将当前中间件应用于 next
func (h *Health) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !h.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		resp := &response{ResponseWriter: w}
		next.ServeHTTP(resp, r)
		go h.save(r.Method, r.URL.Path, time.Now().Sub(start), resp.status)
	})
}

func (h *Health) save(method, path string, dur time.Duration, status int) {
	state := h.store.Get(method, path)

	state.Count++
	state.Last = time.Now()
	state.Spend += dur

	if state.Min > dur {
		state.Min = dur
	} else if state.Max < dur {
		state.Max = dur
	}

	if status >= 400 && status < 500 {
		state.UserErrors++
	} else if status >= 500 {
		state.ServerErrors++
	}

	h.store.Save(state)
}
