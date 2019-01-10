// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package digest

import (
	"io/ioutil"
	"log"
	"net/http"
	"testing"

	"github.com/issue9/assert"
)

var (
	fok = func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}

	hok = http.HandlerFunc(fok)

	authFunc = func(username string) (interface{}, string, bool) {
		return username, username, true
	}
)

func TestNew(t *testing.T) {
	a := assert.New(t)
	var h http.Handler

	a.Panic(func() {
		h = New(nil, nil, "", false, nil)
	})

	a.Panic(func() {
		h = New(hok, nil, "", false, nil)
	})

	a.NotPanic(func() {
		h = New(hok, authFunc, "", false, nil)
	})

	dd, ok := h.(*digest)
	a.True(ok).
		Equal(dd.authorization, "Authorization").
		Equal(dd.authenticate, "WWW-Authenticate").
		Equal(dd.unauthorizationStatus, http.StatusUnauthorized).
		Nil(dd.errlog).
		NotNil(dd.auth)

	a.NotPanic(func() {
		h = New(hok, authFunc, "", true, log.New(ioutil.Discard, "", 0))
	})

	dd, ok = h.(*digest)
	a.True(ok).
		Equal(dd.authorization, "Proxy-Authorization").
		Equal(dd.authenticate, "Proxy-Authenticate").
		Equal(dd.unauthorizationStatus, http.StatusProxyAuthRequired).
		NotNil(dd.errlog).
		NotNil(dd.auth)
}

func TestParseAuthorization(t *testing.T) {
	a := assert.New(t)

	items, err := parseAuthorization("")
	a.Error(err).Nil(items)

	items, err = parseAuthorization("Digest")
	a.Error(err).Nil(items)

	items, err = parseAuthorization("Digest ")
	a.Error(err).Nil(items)

	// 格式不正确
	items, err = parseAuthorization("Digest pop=")
	a.Error(err).Nil(items)

	// 正常
	items, err = parseAuthorization("Digest pop=123")
	a.NotError(err).NotNil(items)

	// 相同名称的键
	items, err = parseAuthorization("Digest pop=123,pop=xxx")
	a.Error(err).Nil(items)

	// 正常
	items, err = parseAuthorization("Digest pop=123,xx=456x")
	a.NotError(err).NotNil(items)
}
