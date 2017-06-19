// Copyright 2017 by caixw, All rights reserved.
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

var _ GenFunc = GenIP

func TestServer_bucket(t *testing.T) {
	a := assert.New(t)
	srv := NewServer(NewMemory(10), 10, 50*time.Second, nil)
	a.NotNil(srv)

	r1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	r1.RemoteAddr = "1"
	r2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	r2.RemoteAddr = "1"

	b1, err := srv.bucket(r1)
	a.NotError(err).NotNil(b1)
	b2, err := srv.bucket(r2)
	a.NotError(err).NotNil(b2)
	a.Equal(b1, b2)

	r2.RemoteAddr = "2"
	b2, err = srv.bucket(r2)
	a.NotError(err).NotNil(b2)
	a.NotEqual(b1, b2)

	r2.RemoteAddr = ""
	b2, err = srv.bucket(r2)
	a.Error(err).Nil(b2)
}

func TestServer_transfer(t *testing.T) {
	a := assert.New(t)
	srv := NewServer(NewMemory(10), 10, 50*time.Second, nil)
	a.NotNil(srv)

	r1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	r1.RemoteAddr = "1"
	r2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	r2.RemoteAddr = "2"

	b1, err := srv.bucket(r1)
	a.NotError(err).NotNil(b1)
	b2, err := srv.bucket(r2)
	a.NotError(err).NotNil(b2)
	a.NotEmpty(b1, b2)

	a.NotError(srv.Transfer("1", "2"))
	b2, err = srv.bucket(r2)
	a.NotError(err).NotNil(b2)
	a.Equal(b1, b2)
}
