// SPDX-License-Identifier: MIT

package digest

import (
	"io/ioutil"
	"log"
	"net/http"
	"testing"

	"github.com/issue9/assert"
)

var (
	authFunc = func(username string) (interface{}, string, bool) {
		return username, username, true
	}
)

func TestNew(t *testing.T) {
	a := assert.New(t)
	var d *Digest

	a.Panic(func() {
		d = New(nil, "", false, nil)
	})

	a.Panic(func() {
		d = New(nil, "", false, nil)
	})

	a.NotPanic(func() {
		d = New(authFunc, "", false, nil)
	})

	a.Equal(d.authorization, "Authorization").
		Equal(d.authenticate, "WWW-Authenticate").
		Equal(d.unauthorizationStatus, http.StatusUnauthorized).
		Nil(d.errlog).
		NotNil(d.auth)

	a.NotPanic(func() {
		d = New(authFunc, "", true, log.New(ioutil.Discard, "", 0))
	})

	a.Equal(d.authorization, "Proxy-Authorization").
		Equal(d.authenticate, "Proxy-Authenticate").
		Equal(d.unauthorizationStatus, http.StatusProxyAuthRequired).
		NotNil(d.errlog).
		NotNil(d.auth)
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
