// SPDX-License-Identifier: MIT

package jwt

import (
	"fmt"
	"io/fs"

	"github.com/golang-jwt/jwt/v4"
	"github.com/issue9/sliceutil"
)

type key struct {
	id              any
	sign            SigningMethod
	private, public any
}

// KeyIDs 所有注册的编码名称
func (j *JWT[T]) KeyIDs() []string { return j.keyIDs }

// AddKey 添加证书
func (j *JWT[T]) AddKey(id string, sign SigningMethod, pvt, pub any) {
	if sliceutil.Exists(j.keys, func(e *key) bool { return e.id == id }) {
		panic(fmt.Sprintf("存在同名的签名方法 %s", id))
	}

	j.keys = append(j.keys, &key{
		id:      id,
		sign:    sign,
		private: pvt,
		public:  pub,
	})
	j.keyIDs = append(j.keyIDs, id)
}

func (j *JWT[T]) AddHMAC(id string, sign *jwt.SigningMethodHMAC, secret []byte) {
	j.AddKey(id, sign, secret, secret)
}

func (j *JWT[T]) addRSA(id string, sign jwt.SigningMethod, private, public []byte) {
	pvt, err := jwt.ParseRSAPrivateKeyFromPEM(private)
	if err != nil {
		panic(err)
	}

	pub, err := jwt.ParseRSAPublicKeyFromPEM(public)
	if err != nil {
		panic(err)
	}

	j.AddKey(id, sign, pvt, pub)
}

func (j *JWT[T]) AddRSA(id string, sign *jwt.SigningMethodRSA, private, public []byte) {
	j.addRSA(id, sign, private, public)
}

func (j *JWT[T]) AddRSAPSS(id string, sign *jwt.SigningMethodRSAPSS, private, public []byte) {
	j.addRSA(id, sign, private, public)
}

func (j *JWT[T]) AddECDSA(id string, sign *jwt.SigningMethodECDSA, private, public []byte) {
	pvt, err := jwt.ParseECPrivateKeyFromPEM(private)
	if err != nil {
		panic(err)
	}

	pub, err := jwt.ParseECPublicKeyFromPEM(public)
	if err != nil {
		panic(err)
	}

	j.AddKey(id, sign, pvt, pub)
}

func (j *JWT[T]) AddEd25519(id string, sign *jwt.SigningMethodEd25519, private, public []byte) {
	pvt, err := jwt.ParseEdPrivateKeyFromPEM(private)
	if err != nil {
		panic(err)
	}

	pub, err := jwt.ParseEdPublicKeyFromPEM(public)
	if err != nil {
		panic(err)
	}

	j.AddKey(id, sign, pvt, pub)
}

func (j *JWT[T]) AddRSAFromFS(id string, sign *jwt.SigningMethodRSA, fsys fs.FS, private, public string) {
	pvt, pub, err := loadFS(fsys, private, public)
	if err != nil {
		panic(err)
	}
	j.AddRSA(id, sign, pvt, pub)
}

func (j *JWT[T]) AddRSAPSSFromFS(id string, sign *jwt.SigningMethodRSAPSS, fsys fs.FS, private, public string) {
	pvt, pub, err := loadFS(fsys, private, public)
	if err != nil {
		panic(err)
	}
	j.AddRSAPSS(id, sign, pvt, pub)
}

func (j *JWT[T]) AddECDSAFromFS(id string, sign *jwt.SigningMethodECDSA, fsys fs.FS, private, public string) {
	pvt, pub, err := loadFS(fsys, private, public)
	if err != nil {
		panic(err)
	}
	j.AddECDSA(id, sign, pvt, pub)
}

func (j *JWT[T]) AddEd25519FromFS(id string, sign *jwt.SigningMethodEd25519, fsys fs.FS, private, public string) {
	pvt, pub, err := loadFS(fsys, private, public)
	if err != nil {
		panic(err)
	}
	j.AddEd25519(id, sign, pvt, pub)
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
