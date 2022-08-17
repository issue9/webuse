// SPDX-License-Identifier: MIT

// Package health API 状态检测
package health

import (
	"time"

	"github.com/issue9/web"
)

// Store 存储 API 状态的接口
type Store interface {
	// Get 获取指定 API 的数据
	//
	// 如果还不存在，则返回空对象。
	Get(method, path string) *State

	// Save 保存数据内容
	//
	// 每生成一条数据，均会以异步的方式调用 Save，由处理具体的操作方式。
	Save(*State)

	// All 返回所有接口的状态信息
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

func newState(method, path string) *State { return &State{Method: method, Path: path} }

// New 声明 Health 实例
func New(store Store) *Health {
	return &Health{
		Enabled: true,
		store:   store,
	}
}

// Register 注册 api
//
// 这不是一个必须的操作，默认情况下，当 api 被第一次访问时，
// 才会将该 api 的信息进行保存，此操作相当于提前进行一次访问。
// 此操作对部分冷门的 api 可以保证其出现在 States() 中。
func (h *Health) Register(method, pattern string) {
	h.store.Save(newState(method, pattern))
}

// States 返回所有的状态列表
func (h *Health) States() []*State { return h.store.All() }

// Middleware 将当前中间件应用于 next
func (h *Health) Middleware(next web.HandlerFunc) web.HandlerFunc {
	return func(ctx *web.Context) web.Responser {
		if !h.Enabled {
			return next(ctx)
		}

		start := time.Now()
		ctx.OnExit(func(status int) {
			req := ctx.Request()
			go h.save(req.Method, ctx.Route().Node().Pattern(), time.Since(start), status)
		})

		return next(ctx)
	}
}

func (h *Health) save(method, path string, dur time.Duration, status int) {
	state := h.store.Get(method, path)

	state.Count++
	state.Last = time.Now()
	state.Spend += dur

	if status >= 400 && status < 500 {
		state.UserErrors++
	} else if status >= 500 {
		state.ServerErrors++
	}

	if state.Count == 1 { // 第一次访问
		state.Min = dur
		state.Max = dur
	} else {
		if state.Min > dur {
			state.Min = dur
		} else if state.Max < dur {
			state.Max = dur
		}
	}

	h.store.Save(state)
}
