// SPDX-License-Identifier: MIT

package host

import "net/http"

type host struct {
	*Host
	next http.Handler
}

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
func (s *Switcher) AddHost(next http.Handler, domain ...string) *Host {
	h := New(false, domain...)
	s.hosts = append(s.hosts, &host{next: next, Host: h})
	return h
}

func (s *Switcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// r.URL.Hostname() 可能是空值
	hostname := r.Host

	for _, host := range s.hosts {
		if host.matched(hostname) {
			host.next.ServeHTTP(w, r)
			return
		}
	}

	http.NotFound(w, r)
}
