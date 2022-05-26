// SPDX-License-Identifier: MIT

package jwt

import (
	"io/fs"

	"github.com/golang-jwt/jwt/v4"
)

type key struct {
	id              any
	sign            SigningMethod
	private, public any
}

// AddKey 添加证书
func (j *JWT[T]) AddKey(id string, sign SigningMethod, pvt, pub any) {
	j.keys = append(j.keys, &key{
		id:      id,
		sign:    sign,
		private: pvt,
		public:  pub,
	})
}

func (j *JWT[T]) AddHMAC(id string, sign *jwt.SigningMethodHMAC, secret []byte) {
	j.AddKey(id, sign, secret, secret)
}

func (j *JWT[T]) addRSA(id string, sign jwt.SigningMethod, private, public []byte) error {
	pvt, err := jwt.ParseRSAPrivateKeyFromPEM(private)
	if err != nil {
		return err
	}

	pub, err := jwt.ParseRSAPublicKeyFromPEM(public)
	if err != nil {
		return err
	}

	j.AddKey(id, sign, pvt, pub)
	return nil
}

func (j *JWT[T]) AddRSA(id string, sign *jwt.SigningMethodRSA, private, public []byte) error {
	return j.addRSA(id, sign, private, public)
}

func (j *JWT[T]) AddRSAPSS(id string, sign *jwt.SigningMethodRSAPSS, private, public []byte) error {
	return j.addRSA(id, sign, private, public)
}

func (j *JWT[T]) AddECDSA(id string, sign *jwt.SigningMethodECDSA, private, public []byte) error {
	pvt, err := jwt.ParseECPrivateKeyFromPEM(private)
	if err != nil {
		return err
	}

	pub, err := jwt.ParseECPublicKeyFromPEM(public)
	if err != nil {
		return err
	}

	j.AddKey(id, sign, pvt, pub)
	return nil
}

func (j *JWT[T]) AddEd25519(id string, sign *jwt.SigningMethodEd25519, private, public []byte) error {
	pvt, err := jwt.ParseEdPrivateKeyFromPEM(private)
	if err != nil {
		return err
	}

	pub, err := jwt.ParseEdPublicKeyFromPEM(public)
	if err != nil {
		return err
	}

	j.AddKey(id, sign, pvt, pub)
	return nil
}

func (j *JWT[T]) AddRSAFromFS(id string, sign *jwt.SigningMethodRSA, fsys fs.FS, private, public string) error {
	pvt, pub, err := loadFS(fsys, private, public)
	if err != nil {
		return err
	}
	return j.AddRSA(id, sign, pvt, pub)
}

func (j *JWT[T]) AddRSAPSSFromFS(id string, sign *jwt.SigningMethodRSAPSS, fsys fs.FS, private, public string) error {
	pvt, pub, err := loadFS(fsys, private, public)
	if err != nil {
		return err
	}
	return j.AddRSAPSS(id, sign, pvt, pub)
}

func (j *JWT[T]) AddECDSAFromFS(id string, sign *jwt.SigningMethodECDSA, fsys fs.FS, private, public string) error {
	pvt, pub, err := loadFS(fsys, private, public)
	if err != nil {
		return err
	}
	return j.AddECDSA(id, sign, pvt, pub)
}

func (j *JWT[T]) AddEd25519FromFS(id string, sign *jwt.SigningMethodEd25519, fsys fs.FS, private, public string) error {
	pvt, pub, err := loadFS(fsys, private, public)
	if err != nil {
		return err
	}
	return j.AddEd25519(id, sign, pvt, pub)
}

func loadFS(fsys fs.FS, private, public string) (pvt []byte, pub []byte, err error) {
	pvt, err = fs.ReadFile(fsys, private)
	if err != nil {
		return nil, nil, err
	}

	pub, err = fs.ReadFile(fsys, public)
	if err != nil {
		return nil, nil, err
	}

	return pvt, pub, nil
}
