// Copyright 2016 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/issue9/assert"
)

var h1 = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(1)
})

func TestRateLimit(t *testing.T) {
	a := assert.New(t)
	srv := NewServer(NewMemory(100), 1, 10*time.Second, GenIP)
	a.NotNil(srv)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test", nil)
	h := srv.RateLimit(h1, nil)
	h.ServeHTTP(w, r)
	a.Equal(w.Code, 1)
	a.Equal(w.Header().Get("X-Rate-Limit-Limit"), "1")
	a.Equal(w.Header().Get("X-Rate-Limit-Remaining"), "0")

	// 没有令牌可用
	w = httptest.NewRecorder()
	r = httptest.NewRequest(http.MethodGet, "/test", nil)
	h = srv.RateLimit(h1, nil)
	h.ServeHTTP(w, r)
	a.Equal(w.Code, http.StatusTooManyRequests)
	a.Equal(w.Header().Get("X-Rate-Limit-Limit"), "1")
	a.Equal(w.Header().Get("X-Rate-Limit-Remaining"), "0")
}
