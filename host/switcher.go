// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package host

import (
	"net/http"
)

// Switcher 实现按域名进行路由
type Switcher struct {
	hosts []*host
}

// NewSwitcher 声明新的 Switcher 实例
func NewSwitcher() *Switcher {
	return &Switcher{
		hosts: make([]*host, 0, 10),
	}
}

// AddHost 添加域名信息
//
// domain 可以是泛域名，比如 *.example.com，但不能是 s1.*.example.com
func (s *Switcher) AddHost(h http.Handler, domain ...string) {
	s.hosts = append(s.hosts, newHost(h, domain...))
}

func (s *Switcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// r.URL.Hostname() 可能是空值
	hostname := r.Host

	for _, host := range s.hosts {
		if host.matched(hostname) {
			host.handler.ServeHTTP(w, r)
			return
		}
	}

	http.NotFound(w, r)
}
