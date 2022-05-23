// SPDX-License-Identifier: MIT

package jwt

import (
	"os"
	"testing"

	"github.com/golang-jwt/jwt/v4"
	"github.com/issue9/assert/v2"
)

func TestNewRSAFromFS(t *testing.T) {
	a := assert.New(t, false)

	j, err := NewRSAFromFS(nil, claimsBuilder, jwt.SigningMethodRS256, os.DirFS("./testdata"), "rsa-private.pem", "rsa-public.pem")
	a.NotError(err).NotNil(j)
	testJWT_Sign(a, j)

	d := &memoryDiscarder{}
	j, err = NewRSAFromFS(d, claimsBuilder, jwt.SigningMethodRS256, os.DirFS("./testdata"), "rsa-private.pem", "rsa-public.pem")
	a.NotError(err).NotNil(j)
	testJWT_Middleware(a, j, d)
}

func TestNewRSAPSSFromFS(t *testing.T) {
	a := assert.New(t, false)

	j, err := NewRSAPSSFromFS(defaultDiscarder{}, claimsBuilder, jwt.SigningMethodPS256, os.DirFS("./testdata"), "rsa-private.pem", "rsa-public.pem")
	a.NotError(err).NotNil(j)
	testJWT_Sign(a, j)

	d := &memoryDiscarder{}
	j, err = NewRSAPSSFromFS(d, claimsBuilder, jwt.SigningMethodPS256, os.DirFS("./testdata"), "rsa-private.pem", "rsa-public.pem")
	a.NotError(err).NotNil(j)
	testJWT_Middleware(a, j, d)
}

func TestNewECDSAFromFS(t *testing.T) {
	a := assert.New(t, false)

	j, err := NewECDSAFromFS(nil, claimsBuilder, jwt.SigningMethodES256, os.DirFS("./testdata"), "ec256-private.pem", "ec256-public.pem")
	a.NotError(err).NotNil(j)
	testJWT_Sign(a, j)

	d := &memoryDiscarder{}
	j, err = NewECDSAFromFS(d, claimsBuilder, jwt.SigningMethodES256, os.DirFS("./testdata"), "ec256-private.pem", "ec256-public.pem")
	a.NotError(err).NotNil(j)
	testJWT_Middleware(a, j, d)
}

func TestNewEd25519FromFS(t *testing.T) {
	a := assert.New(t, false)

	j, err := NewEd25519FromFS(nil, claimsBuilder, jwt.SigningMethodEdDSA, os.DirFS("./testdata"), "ed25519-private.pem", "ed25519-public.pem")
	a.NotError(err).NotNil(j)
	testJWT_Sign(a, j)

	d := &memoryDiscarder{}
	j, err = NewEd25519FromFS(d, claimsBuilder, jwt.SigningMethodEdDSA, os.DirFS("./testdata"), "ed25519-private.pem", "ed25519-public.pem")
	a.NotError(err).NotNil(j)
	testJWT_Middleware(a, j, d)
}
