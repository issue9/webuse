// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package ratelimit 提供了对 X-Rate-Limit 功能的中间件：
//  store := NewMemory(...)
//  srv := NewServer(store)
//  h = srv.RateLimit(h, logs.ERROR())
package ratelimit

import (
	"log"
	"net/http"
)

type rateLimiter struct {
	handler http.Handler
	srv     *Server
	errlog  *log.Logger
}

// RateLimit 限制单一用户的 HTTP 请求数量。会向报头输出以下内容：
//  X-Rate-Limit-Limit: 同一个时间段所允许的请求的最大数目;
//  X-Rate-Limit-Remaining: 在当前时间段内剩余的请求的数量;
//  X-Rate-Limit-Reset: 为了得到最大请求数所等待的秒数。
func (srv *Server) RateLimit(h http.Handler, errlog *log.Logger) http.Handler {
	return &rateLimiter{
		handler: h,
		srv:     srv,
		errlog:  errlog,
	}
}

// RateLimitFunc 限制单一用户的 HTTP 请求数量。
func (srv *Server) RateLimitFunc(f func(w http.ResponseWriter, r *http.Request), errlog *log.Logger) http.Handler {
	return srv.RateLimit(http.HandlerFunc(f), errlog)
}

func (l *rateLimiter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b, allow, err := l.srv.allow(r)

	if err != nil {
		if l.errlog != nil {
			l.errlog.Println(err)
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	b.setHeader(w)

	if allow {
		l.handler.ServeHTTP(w, r)
		return
	}
}
