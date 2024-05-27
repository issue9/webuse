// SPDX-FileCopyrightText: 2015-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package health API 状态统计
package health

import (
	"net/http"
	"time"

	"github.com/issue9/web"
)

// State 实际存在的数据类型
type State struct {
	XMLName      struct{}      `xml:"state" yaml:"-" json:"-" cbor:"-"`
	Router       string        `xml:"router" yaml:"router" json:"router" cbor:"router"`                              // 多个路由时，表示的路由名称
	Method       string        `xml:"method,attr" yaml:"method" json:"method" cbor:"method"`                         // 请求方法
	Pattern      string        `xml:"pattern" yaml:"pattern" json:"pattern" cbor:"pattern"`                          // 路由
	Min          time.Duration `xml:"min,attr" yaml:"min" json:"min" cbor:"min"`                                     // 最小的执行时间
	Max          time.Duration `xml:"max,attr" yaml:"max" json:"max" cbor:"max"`                                     // 最大的执行时间
	Count        int           `xml:"count,attr" yaml:"count" json:"count" cbor:"count"`                             // 总的请求次数
	UserErrors   int           `xml:"userErrors,attr" yaml:"userErrors" json:"userErrors" cbor:"userErrors"`         // 用户端出错次数，400-499
	ServerErrors int           `xml:"serverErrors,attr" yaml:"serverErrors" json:"serverErrors" cbor:"serverErrors"` // 服务端出错次数，>500
	Last         time.Time     `xml:"last" yaml:"last" json:"last" cbor:"last"`                                      // 最后的访问时间
	Spend        time.Duration `xml:"spend,attr" yaml:"spend" json:"spend" cbor:"spend"`                             // 总花费的时间
}

// Health API 状态检测
type Health struct {
	Enabled bool // 是否启用当前的中间件
	store   Store
}

func newState(route, method, path string) *State {
	return &State{Router: route, Method: method, Pattern: path}
}

// New 声明 [Health] 实例
func New(store Store) *Health { return &Health{Enabled: true, store: store} }

func (h *Health) Plugin(s web.Server) {
	s.Routers().Use(web.MiddlewareFunc(func(next web.HandlerFunc, method, pattern, router string) web.HandlerFunc {
		if method != http.MethodOptions && method != "" && method != http.MethodHead &&
			pattern != "" &&
			h.store.Get(router, method, pattern) == nil {
			h.store.Save(newState(router, method, pattern))
		}
		return next
	}))

	s.OnExitContext(func(ctx *web.Context, status int) {
		if h.Enabled {
			h.save(ctx, status)
		}
	})
}

// States 返回所有的状态列表
func (h *Health) States() []*State { return h.store.All() }

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
		state.Min = min(state.Min, dur)
		state.Max = max(state.Max, dur)
	}

	h.store.Save(state)
}
