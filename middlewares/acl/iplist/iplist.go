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

// IPLister 根据客户端的 IP 过滤
type IPLister interface {
	web.Middleware

	// Set 设置名单列表
	//
	// ip IP 地址，可以是具体的 IP 地址，比如 44.44.44.44，
	// 也可以是通配符(以 /* 结尾)，比如 2.2/*，表示匹配 2.2.1.1 也匹配 2.22.1.1，
	// 但是并不匹配 ipv6 的相同地址 ::0202:0202。
	// 如果已经存在相同的值，则不会重复添加。
	//
	// NOTE: 传递空值将清空内容。
	Set(ip ...string)

	// List 返回名单列表
	List() []string
}

type common struct {
	list     []string
	wildcard []string
}

type white struct {
	*common
}

type black struct {
	*common
}

func newCommon() *common {
	return &common{
		list:     make([]string, 0, 10),
		wildcard: make([]string, 0, 10),
	}
}

// NewWhite 声明白名单过滤器
//
// 所有未在白名单中的 IP 都将被禁止访问。
func NewWhite() IPLister { return &white{common: newCommon()} }

// NewWhite 声明黑名单过滤器
//
// 所有未在黑名单中的 IP 才允许访问。
func NewBlack() IPLister { return &black{common: newCommon()} }

func (l *common) Set(ip ...string) {
	for _, i := range ip {
		if strings.HasSuffix(i, "/*") {
			i = strings.TrimSuffix(i, "/*")

			if slices.Index(l.wildcard, i) < 0 {
				l.wildcard = append(l.wildcard, i)
			}
		} else {
			if slices.Index(l.list, i) < 0 {
				l.list = append(l.list, i)
			}
		}
	}
}

func (l *common) List() []string {
	list := slices.Clone(l.list)
	for _, w := range l.wildcard {
		list = append(list, w+"/*")
	}
	return list
}

func (l *common) match(ip string) bool {
	if slices.Index(l.list, ip) >= 0 {
		return true
	}

	for _, ip := range l.wildcard {
		if strings.HasPrefix(ip, ip) {
			return true
		}
	}

	return false
}

func (l *white) Middleware(next web.HandlerFunc, _, _, _ string) web.HandlerFunc {
	return func(ctx *web.Context) web.Responser {
		host, err := splitIP(ctx.ClientIP())
		if err != nil {
			return ctx.Error(err, web.ProblemBadRequest)
		}

		if l.match(host) {
			return next(ctx)
		}
		return ctx.Problem(web.ProblemForbidden)
	}
}

func (l *black) Middleware(next web.HandlerFunc, _, _, _ string) web.HandlerFunc {
	return func(ctx *web.Context) web.Responser {
		host, err := splitIP(ctx.ClientIP())
		if err != nil {
			return ctx.Problem(web.ProblemBadRequest)
		}

		if l.match(host) {
			return ctx.Problem(web.ProblemForbidden)
		}
		return next(ctx)
	}
}

func splitIP(ip string) (string, error) {
	if ip[0] == '[' { // ipv6 且带端口
		if ip[len(ip)-1] == ']' {
			return ip[1 : len(ip)-1], nil
		} else if index := strings.LastIndex(ip, "]:"); index >= 0 {
			return ip[1:index], nil
		}
		return "", web.NewLocaleError("invalid ip %s", ip)
	}

	if index := strings.LastIndexByte(ip, ':'); index >= 0 {
		ip4 := ip[:index]
		if strings.IndexByte(ip4, ':') >= 0 { // ip4 不可能包含两个 :
			return "", web.NewLocaleError("invalid ip %s", ip)
		}
		return ip4, nil
	}
	return ip, nil
}
