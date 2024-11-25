// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package token

// Response 申请令牌时返回的对象
type Response struct {
	XMLName      struct{} `json:"-" cbor:"-" xml:"token"`
	AccessToken  string   `json:"access_token" xml:"access_token" cbor:"access_token" comment:"access token"`            // 访问令牌
	RefreshToken string   `json:"refresh_token" xml:"refresh_token" cbor:"refresh_token" comment:"refresh token"`        // 刷新令牌
	AccessExp    int      `json:"access_exp" xml:"access_exp,attr" cbor:"access_exp" comment:"access token expired"`     // 访问令牌的有效时长，单位为秒
	RefreshExp   int      `json:"refresh_exp" xml:"refresh_exp,attr" cbor:"refresh_exp" comment:"refresh token expired"` // 刷新令牌的有效时长，单位为秒
}

// BuildResponseFunc 构建将令牌返回给客户的结构体
//
// access 访问令牌是必须的；
// refresh 刷新令牌；
// accessExp 表示访问令牌的过期时间；
// refreshExp 表示刷新令牌的过期时间；
//
// NOTE: 之所以定义此方法，主要是用于让部分未内置在框架中的 mimetype 也可以方便处理。
type BuildResponseFunc = func(access, refresh string, accessExp, refreshExp int) any

// DefaultBuildResponse [BuildResponseFunc] 的默认实现
func DefaultBuildResponse(access, refresh string, ae, re int) any {
	return &Response{AccessToken: access, RefreshToken: refresh, AccessExp: ae, RefreshExp: re}
}
