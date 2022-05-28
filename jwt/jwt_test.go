// SPDX-License-Identifier: MIT

package jwt

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/issue9/assert/v2"
	"github.com/issue9/web"
	"github.com/issue9/web/server"
	"github.com/issue9/web/server/servertest"
)

func newJWT(a *assert.Assertion, expired time.Duration) (*stdSigner, *stdVerifier, *memoryBlocker) {
	s := NewSigner[*jwt.RegisteredClaims](expired)
	a.NotNil(s)

	m := &memoryBlocker{}
	j := NewVerifier[*jwt.RegisteredClaims](m, func() *jwt.RegisteredClaims {
		return &jwt.RegisteredClaims{}
	})
	a.NotNil(j)

	return s, j, m
}

func TestVerifier_Middleware(t *testing.T) {
	a := assert.New(t, false)

	signer, verifier, m := newJWT(a, time.Hour)
	signer.AddHMAC("hmac-secret", jwt.SigningMethodHS256, []byte("secret"))
	verifier.AddHMAC("hmac-secret", jwt.SigningMethodHS256, []byte("secret"))
	verifierMiddleware(a, signer, verifier, m)

	a.PanicString(func() {
		verifier.AddHMAC("hmac-secret", jwt.SigningMethodHS256, []byte("secret"))
	}, "存在同名的签名方法 hmac-secret")

	a.PanicString(func() {
		signer.AddHMAC("hmac-secret", jwt.SigningMethodHS256, []byte("secret"))
	}, "存在同名的签名方法 hmac-secret")

	signer, verifier, m = newJWT(a, time.Hour)
	signer.AddRSAFromFS("rsa", jwt.SigningMethodRS256, os.DirFS("./testdata"), "rsa-private.pem")
	verifier.AddRSAFromFS("rsa", jwt.SigningMethodRS256, os.DirFS("./testdata"), "rsa-public.pem")
	verifierMiddleware(a, signer, verifier, m)

	signer, verifier, m = newJWT(a, time.Hour)
	signer.AddRSAPSSFromFS("rsa-pss", jwt.SigningMethodPS256, os.DirFS("./testdata"), "rsa-private.pem")
	verifier.AddRSAPSSFromFS("rsa-pss", jwt.SigningMethodPS256, os.DirFS("./testdata"), "rsa-public.pem")
	verifierMiddleware(a, signer, verifier, m)

	signer, verifier, m = newJWT(a, time.Hour)
	signer.AddECDSAFromFS("ecdsa", jwt.SigningMethodES256, os.DirFS("./testdata"), "ec256-private.pem")
	verifier.AddECDSAFromFS("ecdsa", jwt.SigningMethodES256, os.DirFS("./testdata"), "ec256-public.pem")
	verifierMiddleware(a, signer, verifier, m)

	signer, verifier, m = newJWT(a, time.Hour)
	signer.AddEd25519FromFS("ed25519", jwt.SigningMethodEdDSA, os.DirFS("./testdata"), "ed25519-private.pem")
	verifier.AddEd25519FromFS("ed25519", jwt.SigningMethodEdDSA, os.DirFS("./testdata"), "ed25519-public.pem")
	verifierMiddleware(a, signer, verifier, m)

	signer, verifier, m = newJWT(a, time.Hour)
	signer.AddEd25519FromFS("ed25519", jwt.SigningMethodEdDSA, os.DirFS("./testdata"), "ed25519-private.pem")
	verifier.AddEd25519FromFS("ed25519", jwt.SigningMethodEdDSA, os.DirFS("./testdata"), "ed25519-public.pem")
	signer.AddECDSAFromFS("ecdsa", jwt.SigningMethodES256, os.DirFS("./testdata"), "ec256-private.pem")
	verifier.AddECDSAFromFS("ecdsa", jwt.SigningMethodES256, os.DirFS("./testdata"), "ec256-public.pem")
	signer.AddRSAPSSFromFS("rsa-pss", jwt.SigningMethodPS256, os.DirFS("./testdata"), "rsa-private.pem")
	verifier.AddRSAPSSFromFS("rsa-pss", jwt.SigningMethodPS256, os.DirFS("./testdata"), "rsa-public.pem")
	signer.AddRSAFromFS("rsa", jwt.SigningMethodRS256, os.DirFS("./testdata"), "rsa-private.pem")
	verifier.AddRSAFromFS("rsa", jwt.SigningMethodRS256, os.DirFS("./testdata"), "rsa-public.pem")
	verifierMiddleware(a, signer, verifier, m)
	verifierMiddleware(a, signer, verifier, m)
	verifierMiddleware(a, signer, verifier, m)
	verifierMiddleware(a, signer, verifier, m)
}

func verifierMiddleware(a *assert.Assertion, signer *stdSigner, verifier *stdVerifier, d *memoryBlocker) {
	a.TB().Helper()
	d.clear()

	claims := &jwt.RegisteredClaims{
		Issuer:  "issuer",
		Subject: "subject",
		ID:      "id",
	}

	s := servertest.NewTester(a, nil)
	r := s.NewRouter()
	r.Post("/login", func(ctx *web.Context) web.Responser {
		return signer.RenderAccess(ctx, http.StatusCreated, &Response{}, claims)
	})

	r.Get("/info", verifier.Middleware(func(ctx *server.Context) server.Responser {
		val, found := verifier.GetValue(ctx)
		if !found {
			return ctx.Status(http.StatusNotFound)
		}

		if val.Issuer != claims.Issuer || val.Subject != claims.Subject || val.ID != claims.ID {
			return ctx.Status(http.StatusUnauthorized)
		}

		return ctx.OK(nil)
	}))

	r.Delete("/login", verifier.Middleware(func(ctx *web.Context) web.Responser {
		val, found := verifier.GetValue(ctx)
		if !found {
			return ctx.Status(http.StatusNotFound)
		}

		if val.Issuer != claims.Issuer || val.Subject != claims.Subject || val.ID != claims.ID {
			return ctx.Status(http.StatusUnauthorized)
		}

		if d != nil {
			d.BlockToken(verifier.GetToken(ctx))
		}

		return ctx.NoContent()
	}))

	s.GoServe()

	s.NewRequest(http.MethodPost, "/login", nil).
		Do(nil).
		Status(http.StatusCreated).BodyFunc(func(a *assert.Assertion, body []byte) {
		resp := &Response{}
		a.NotError(json.Unmarshal(body, resp))
		a.NotEmpty(resp).
			NotEmpty(resp.Access).
			Zero(resp.Refresh)

		s.Get("/info").
			Header("Authorization", "BEARER "+resp.Access).
			Do(nil).
			Status(http.StatusOK)

		s.Delete("/login").
			Header("Authorization", resp.Access).
			Do(nil).
			Status(http.StatusNoContent)

		// token 已经在 delete /login 中被弃用
		s.Get("/info").
			Header("Authorization", resp.Access).
			Do(nil).
			Status(http.StatusUnauthorized)
	})

	s.Close(0)
	s.Wait()
}

func TestVerifier_client(t *testing.T) {
	a := assert.New(t, false)
	signer, verifier, _ := newJWT(a, time.Hour)
	signer.AddRSAFromFS("rsa", jwt.SigningMethodRS256, os.DirFS("./testdata"), "rsa-private.pem")
	verifier.AddRSAFromFS("rsa", jwt.SigningMethodRS256, os.DirFS("./testdata"), "rsa-public.pem")

	claims := &jwt.RegisteredClaims{
		Issuer:  "issuer",
		Subject: "subject",
		ID:      "id",
	}

	s := servertest.NewTester(a, nil)
	r := s.NewRouter()
	r.Post("/login", func(ctx *web.Context) web.Responser {
		return signer.RenderAccessRefresh(ctx, http.StatusCreated, &Response{}, claims, claims)
	})

	r.Get("/info", verifier.Middleware(func(ctx *server.Context) server.Responser {
		val, found := verifier.GetValue(ctx)
		if !found {
			return ctx.Status(http.StatusNotFound)
		}

		if val.Issuer != claims.Issuer || val.Subject != claims.Subject || val.ID != claims.ID {
			return ctx.Status(http.StatusUnauthorized)
		}

		return ctx.OK(nil)
	}))

	s.GoServe()

	s.NewRequest(http.MethodPost, "/login", nil).
		Do(nil).
		Status(http.StatusCreated).BodyFunc(func(a *assert.Assertion, body []byte) {
		m := &Response{}
		a.NotError(json.Unmarshal(body, &m))
		a.NotEmpty(m).
			NotEmpty(m.Access).
			NotEmpty(m.Refresh)

		token, parts, err := jwt.NewParser().ParseUnverified(m.Access, &jwt.RegisteredClaims{})
		a.NotError(err).Equal(3, len(parts)).NotNil(token)
		header := decodeHeader(a, parts[0])
		a.Equal(header["alg"], "none").NotEmpty(header["kid"])

		// 改变 alg，影响
		header["alg"] = "ES256"
		parts[0] = encodeHeader(a, header)
		s.Get("/info").
			Header("Authorization", "BEARER "+strings.Join(parts, ".")).
			Do(nil).
			Status(http.StatusInternalServerError)

		// 改变 kid(kid 存在)，影响
		signer.AddECDSAFromFS("ecdsa", jwt.SigningMethodES256, os.DirFS("./testdata"), "ec256-private.pem")
		verifier.AddECDSAFromFS("ecdsa", jwt.SigningMethodES256, os.DirFS("./testdata"), "ec256-public.pem")
		header["kid"] = "ecdsa"
		header["alg"] = "none"
		parts[0] = encodeHeader(a, header)
		s.Get("/info").
			Header("Authorization", "BEARER "+strings.Join(parts, ".")).
			Do(nil).
			Status(http.StatusInternalServerError)

		// 改变 kid(kid 不存在)，影响
		header["kid"] = "not-exists"
		header["alg"] = "none"
		parts[0] = encodeHeader(a, header)
		s.Get("/info").
			Header("Authorization", "BEARER "+strings.Join(parts, ".")).
			Do(nil).
			Status(http.StatusInternalServerError)
	})

	s.Close(0)
	s.Wait()
}

func encodeHeader(a *assert.Assertion, header map[string]any) string {
	a.TB().Helper()

	data, err := json.Marshal(header)
	a.NotError(err).NotNil(data)

	return base64.RawURLEncoding.EncodeToString(data)
}

func decodeHeader(a *assert.Assertion, seg string) map[string]any {
	a.TB().Helper()

	var data []byte
	var err error
	if jwt.DecodePaddingAllowed {
		if l := len(seg) % 4; l > 0 {
			seg += strings.Repeat("=", 4-l)
		}
		data, err = base64.URLEncoding.DecodeString(seg)
	} else {
		data, err = base64.RawURLEncoding.DecodeString(seg)
	}
	a.NotError(err).NotNil(data)

	header := make(map[string]any, 3)
	err = json.Unmarshal(data, &header)
	a.NotError(err)

	return header
}
