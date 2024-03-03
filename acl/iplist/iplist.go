// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

// Package iplist 根据 IP 对请求进行过滤
package iplist

import (
	"slices"
	"strings"

	"github.com/issue9/web"
)

type IPList struct {
	enable bool

	white         []string
	whiteWildcard []string

	black         []string
	blackWildcard []string
}

// New 声明 [IPList] 对象
//
// white 和 black 分别表示白名单和黑名单的值，格式可参考 [IPList.WithWhite]。
// 如果同时存在于黑名单和白名单，则白名单优先于黑名单。
func New(white, black []string) *IPList {
	return (&IPList{enable: true}).WithBlack(black...).WithWhite(white...)
}

func (l *IPList) Enable(v bool) {
	l.enable = v
}

// WithWhite 添加白名单
//
// ip IP 地址，可以是具体的 IP 地址，比如 44.44.44.44，
// 也可以是通配符，比如 44.44/*。
func (l *IPList) WithWhite(ip ...string) *IPList {
	for _, i := range ip {
		if strings.HasSuffix(i, "/*") {
			i = strings.TrimSuffix(i, "/*")

			if slices.Index(l.whiteWildcard, i) < 0 {
				l.whiteWildcard = append(l.whiteWildcard, i)
			}
		} else {
			if slices.Index(l.white, i) < 0 {
				l.white = append(l.white, i)
			}
		}
	}

	return l
}

// WithBlack 添加黑名单
//
// ip IP 地址，格式可参考 [IPList.WithWhite]。
func (l *IPList) WithBlack(ip ...string) *IPList {
	for _, i := range ip {
		if strings.HasSuffix(i, "/*") {
			i = strings.TrimSuffix(i, "/*")
			if slices.Index(l.blackWildcard, i) < 0 {
				l.blackWildcard = append(l.blackWildcard, i)
			}
		} else {
			if slices.Index(l.black, i) < 0 {
				l.black = append(l.black, i)
			}
		}
	}

	return l
}

func (l *IPList) Middleware(next web.HandlerFunc) web.HandlerFunc {
	return func(ctx *web.Context) web.Responser {
		if !l.enable {
			return next(ctx)
		}

		cip := ctx.ClientIP()

		if slices.Index(l.white, cip) >= 0 {
			return next(ctx)
		}

		for _, ip := range l.whiteWildcard {
			if strings.HasPrefix(cip, ip) {
				return next(ctx)
			}
		}

		if slices.Index(l.black, cip) >= 0 {
			return ctx.Problem(web.ProblemForbidden)
		}

		for _, ip := range l.blackWildcard {
			if strings.HasPrefix(cip, ip) {
				return ctx.Problem(web.ProblemForbidden)
			}
		}

		return next(ctx)
	}
}
