// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import "net/http"

type host struct {
	domains []string
	handler http.Handler
}

// Host 声明一个限定域名的 handler
func Host(h http.Handler, domains ...string) *host {
	return &host{
		domains: domains,
		handler: h,
	}
}

// HostFunc 将一个 http.HandlerFunc 包装成 http.Handler
func HostFunc(f func(http.ResponseWriter, *http.Request), domains ...string) *host {
	return Host(http.HandlerFunc(f), domains...)
}

func (h *host) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, domain := range h.domains {
		if domain == r.URL.Host {
			h.handler.ServeHTTP(w, r)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
}
