// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"log"
	"net/http"

	"github.com/issue9/handlers/ratelimit"
)

type rateLimiter struct {
	handler http.Handler
	srv     ratelimit.Server
	errlog  *log.Logger
}

// RateLimit 限制单一用户的 HTTP 请求数量。会向报头输出以下内容：
//
// X-Rate-Limit-Limit: 同一个时间段所允许的请求的最大数目;
// X-Rate-Limit-Remaining: 在当前时间段内剩余的请求的数量;
// X-Rate-Limit-Reset: 为了得到最大请求数所等待的秒数。
func RateLimit(h http.Handler, srv ratelimit.Server, errlog *log.Logger) http.Handler {
	return &rateLimiter{
		handler: h,
		srv:     srv,
		errlog:  errlog,
	}
}

// RateLimitFunc 限制单一用户的 HTTP 请求数量。
func RateLimitFunc(f func(w http.ResponseWriter, r *http.Request), srv ratelimit.Server, errlog *log.Logger) http.Handler {
	return RateLimit(http.HandlerFunc(f), srv, errlog)
}

func (l *rateLimiter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b, allow, err := l.srv.Allow(r)

	if err != nil {
		if l.errlog != nil {
			l.errlog.Println(err)
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	b.SetHeader(w)

	if allow {
		l.handler.ServeHTTP(w, r)
		return
	}
}
