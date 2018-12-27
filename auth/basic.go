// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package auth

import (
	"encoding/base64"
	"log"
	"net/http"
	"strings"
)

type basic struct {
	next   http.Handler
	secret string
	realm  string
	errlog *log.Logger

	authorization string
	authenticate  string
}

// NewBasic 新的 Basic 验证方式
func NewBasic(next http.Handler, username, password, realm string, proxy bool, log *log.Logger) http.Handler {
	authorization := "Authorization"
	authenticate := "WWW-Authenticate"
	if proxy {
		authorization = "Proxy-Authorization"
		authenticate = "Proxy-Authenticate"
	}

	return &basic{
		next:   next,
		secret: base64.StdEncoding.EncodeToString([]byte(username + ":" + password)),
		realm:  `Basic realm="` + realm + `"`,
		errlog: log,

		authorization: authorization,
		authenticate:  authenticate,
	}
}

func (b *basic) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get(b.authorization)
	index := strings.IndexByte(auth, ' ')

	if index <= 0 || index >= len(auth) || auth[:index] != "Basic" {
		b.unauthorization(w)
		return
	}

	secret, err := base64.StdEncoding.DecodeString(auth[index+1:])
	if err != nil {
		b.errlog.Println(err)
		b.unauthorization(w)
		return
	}

	if string(secret) != b.secret {
		b.unauthorization(w)
		return
	}

	b.next.ServeHTTP(w, r)
}

func (b *basic) unauthorization(w http.ResponseWriter) {
	w.Header().Set(b.authenticate, b.realm)
	http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
}
