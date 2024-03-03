// SPDX-FileCopyrightText: 2015-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package health API 状态检测
package health

import (
	"net/http"
	"time"

	"github.com/issue9/web"
)

// Store 存储 API 状态的接口
type Store interface {
	// Get 获取指定 API 的数据
	//
	// 如果还不存在，则返回空对象。
	Get(route, method, pattern string) *State

	// Save 保存数据内容
	Save(*State)

	// All 返回所有接口的状态信息
	All() []*State
}

// State 实际存在的数据类型
type State struct {
	XMLName      struct{}      `xml:"state" yaml:"-" json:"-"`
	Route        string        `xml:"route" yaml:"route" json:"route"`         // 多个路由时，表示的路由名称
	Method       string        `xml:"method,attr" yaml:"method" json:"method"` // 请求方法
	Pattern      string        `xml:"pattern" yaml:"pattern" json:"pattern"`   // 路由
	Min          time.Duration `xml:"min,attr" yaml:"min" json:"min"`
	Max          time.Duration `xml:"max,attr" yaml:"max" json:"max"`
	Count        int           `xml:"count,attr" yaml:"count" json:"count"`                      // 总的请求次数
	UserErrors   int           `xml:"userErrors,attr" yaml:"userErrors" json:"userErrors"`       // 用户端出错次数，400-499
	ServerErrors int           `xml:"serverErrors,attr" yaml:"serverErrors" json:"serverErrors"` // 服务端出错次数，>500
	Last         time.Time     `xml:"last" yaml:"last" json:"last"`                              // 最后的访问时间
	Spend        time.Duration `xml:"spend,attr" yaml:"spend" json:"spend"`                      // 总花费的时间
}

// Health API 状态检测
type Health struct {
	Enabled bool // 是否启用当前的中间件
	store   Store
}

func newState(route, method, path string) *State {
	return &State{Route: route, Method: method, Pattern: path}
}

// New 声明 [Health] 实例
func New(store Store) *Health {
	return &Health{Enabled: true, store: store}
}

// Register 注册 api
//
// 这不是一个必须的操作，默认情况下，当 api 被第一次访问时，
// 才会将该 api 的信息进行保存，此操作相当于提前进行一次访问。
// 此操作对部分冷门的 api 可以保证其出现在 States() 中。
func (h *Health) Register(route, method, pattern string) {
	h.store.Save(newState(route, method, pattern))
}

// States 返回所有的状态列表
func (h *Health) States() []*State { return h.store.All() }

// Middleware 将当前中间件应用于 next
func (h *Health) Middleware(next web.HandlerFunc) web.HandlerFunc {
	return func(ctx *web.Context) web.Responser {
		if !h.Enabled {
			return next(ctx)
		}

		ctx.OnExit(func(c *web.Context, status int) {
			h.save(c, status)
		})

		return next(ctx)
	}
}

func (h *Health) save(ctx *web.Context, status int) {
	route := ctx.Route()
	state := h.store.Get(route.RouterName(), ctx.Request().Method, route.Node().Pattern())

	dur := time.Since(ctx.Begin())
	state.Count++
	state.Last = ctx.Begin()
	state.Spend += dur

	if status >= http.StatusBadRequest && status < http.StatusInternalServerError {
		state.UserErrors++
	} else if status >= http.StatusInternalServerError {
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
