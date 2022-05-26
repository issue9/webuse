// SPDX-License-Identifier: MIT

package jwt

import (
	"os"
	"testing"

	"github.com/golang-jwt/jwt/v4"
	"github.com/issue9/assert/v2"
)

func TestJWT_AddHMAC(t *testing.T) {
	a := assert.New(t, false)
	j, m := newJWT(a)

	j.AddHMAC("hmac-secret", jwt.SigningMethodHS256, []byte("secret"))
	testJWT_Middleware(a, j, m, "hmac-secret")
}

func TestJWT_AddFS(t *testing.T) {
	a := assert.New(t, false)
	j, m := newJWT(a)

	err := j.AddRSAFromFS("rsa", jwt.SigningMethodRS256, os.DirFS("./testdata"), "rsa-private.pem", "rsa-public.pem")
	a.NotError(err)
	testJWT_Middleware(a, j, m, "rsa")

	err = j.AddRSAPSSFromFS("rsa-pss", jwt.SigningMethodPS256, os.DirFS("./testdata"), "rsa-private.pem", "rsa-public.pem")
	a.NotError(err)

	err = j.AddECDSAFromFS("ecdsa", jwt.SigningMethodES256, os.DirFS("./testdata"), "ec256-private.pem", "ec256-public.pem")
	a.NotError(err)

	err = j.AddEd25519FromFS("ed25519", jwt.SigningMethodEdDSA, os.DirFS("./testdata"), "ed25519-private.pem", "ed25519-public.pem")
	a.NotError(err)

	testJWT_Middleware(a, j, m, "rsa-pss")
	testJWT_Middleware(a, j, m, "ecdsa")
	testJWT_Middleware(a, j, m, "ed25519")
}
