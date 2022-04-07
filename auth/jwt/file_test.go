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

	j, err := NewRSAFromFS(claimsBuilder, jwt.SigningMethodRS256, os.DirFS("./testdata"), "rsa-private.pem", "rsa-public.pem")
	a.NotError(err).NotNil(j)
	testJWT_Sign(a, j)

	j, err = NewRSAFromFS(claimsBuilder, jwt.SigningMethodRS256, os.DirFS("./testdata"), "rsa-private.pem", "rsa-public.pem")
	a.NotError(err).NotNil(j)
	testJWT_Middleware(a, j)
}

func TestNewRSAPSSFromFS(t *testing.T) {
	a := assert.New(t, false)

	j, err := NewRSAPSSFromFS(claimsBuilder, jwt.SigningMethodPS256, os.DirFS("./testdata"), "rsa-private.pem", "rsa-public.pem")
	a.NotError(err).NotNil(j)
	testJWT_Sign(a, j)

	j, err = NewRSAPSSFromFS(claimsBuilder, jwt.SigningMethodPS256, os.DirFS("./testdata"), "rsa-private.pem", "rsa-public.pem")
	a.NotError(err).NotNil(j)
	testJWT_Middleware(a, j)
}

func TestNewECDSAFromFS(t *testing.T) {
	a := assert.New(t, false)

	j, err := NewECDSAFromFS(claimsBuilder, jwt.SigningMethodES256, os.DirFS("./testdata"), "ec256-private.pem", "ec256-public.pem")
	a.NotError(err).NotNil(j)
	testJWT_Sign(a, j)

	j, err = NewECDSAFromFS(claimsBuilder, jwt.SigningMethodES256, os.DirFS("./testdata"), "ec256-private.pem", "ec256-public.pem")
	a.NotError(err).NotNil(j)
	testJWT_Middleware(a, j)
}

func TestNewEd25519FromFS(t *testing.T) {
	a := assert.New(t, false)

	j, err := NewEd25519FromFS(claimsBuilder, jwt.SigningMethodEdDSA, os.DirFS("./testdata"), "ed25519-private.pem", "ed25519-public.pem")
	a.NotError(err).NotNil(j)
	testJWT_Sign(a, j)

	j, err = NewEd25519FromFS(claimsBuilder, jwt.SigningMethodEdDSA, os.DirFS("./testdata"), "ed25519-private.pem", "ed25519-public.pem")
	a.NotError(err).NotNil(j)
	testJWT_Middleware(a, j)
}
