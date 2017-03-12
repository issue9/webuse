// Copyright 2017 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import "net/http"

type host struct {
	domain  string
	handler http.Handler
}

func (h *host) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.domain != r.URL.Host {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	h.handler.ServeHTTP(w, r)
}

// Host 声明一个限定域名的 handler
func Host(h http.Handler, domain string) *host {
	return &host{
		domain:  domain,
		handler: h,
	}
}

// HostFunc 将一个 http.HandlerFunc 包装成 http.Handler
func HostFunc(f func(http.ResponseWriter, *http.Request), domain string) *host {
	return Host(http.HandlerFunc(f), domain)
}
