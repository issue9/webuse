// SPDX-FileCopyrightText: 2022-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package gb11643 解析身分证详情
package gb11643

import (
	"time"

	"github.com/issue9/web/locales"
)

// 我国现行的身份证号码有两种标准：GB11643-1989、GB11643-1999：
//
// GB11643-1989 为一代身份证，从左至右分别为：
//  ------------------------------------------------------------
//  | 6 位行政区域代码 | 6 位出生年日期（不含世纪数）| 3 位顺序码 |
//  ------------------------------------------------------------
//
// GB11643-1999 为二代身份证，从左至右分别为：
//  ------------------------------------------------------------
//  | 6 位行政区域代码 |  8 位出生日期 |  3 位顺序码 |  1 位检验码 |
//  ------------------------------------------------------------

const layout = "20060102"

// GB11643 身份证信息
type GB11643 struct {
	Raw    string    // 原始数据
	Region string    // 区域代码
	Date   time.Time // 出生年月
	IsMale bool      // true 为男姓，false 为女生
}

// Parse 分析身份证信息
func Parse(bs string) (*GB11643, error) {
	if !IsValid([]byte(bs)) {
		return nil, locales.ErrInvalidFormat()
	}

	switch len(bs) {
	case 15:
		return parse15(bs)
	case 18:
		return parse18(bs)
	default:
		return nil, locales.ErrInvalidFormat()
	}
}

func parse15(bs string) (*GB11643, error) {
	date, err := time.Parse(layout, "19"+bs[6:12])
	if err != nil {
		return nil, err
	}

	return &GB11643{
		Raw:    bs,
		Region: bs[:6],
		Date:   date,
		IsMale: (bs[14]-'0')%2 == 1,
	}, nil
}

func parse18(bs string) (*GB11643, error) {
	date, err := time.Parse(layout, bs[6:14])
	if err != nil {
		return nil, err
	}

	return &GB11643{
		Raw:    bs,
		Region: bs[:6],
		Date:   date,
		IsMale: (bs[16]-'0')%2 == 1,
	}, nil
}
