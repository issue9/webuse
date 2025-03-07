// SPDX-FileCopyrightText: 2022-2025 caixw
//
// SPDX-License-Identifier: MIT

package validator

import (
	"net/mail"
	"net/netip"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/text/language"

	"github.com/issue9/webuse/v7/filters/gb11643"
	"github.com/issue9/webuse/v7/filters/gb32100"
	"github.com/issue9/webuse/v7/filters/luhn"
	"github.com/issue9/webuse/v7/filters/strength"
	"github.com/issue9/webuse/v7/internal/isbn"
)

// Strength 声明密码强度的验证对象
//
// length 对长度的最小要求；
// upper 对大写字符的最小要求；
// lower 对小写字符的最小要求；
// punct 对符号的最小要求；
// num 对数字的最小要求
//
// 详细文档可参考 [strength.Strength]。
func Strength(length, upper, lower, punct, num int8) func(string) bool {
	return (&strength.Strength{Length: length, Upper: upper, Lower: lower, Punct: punct, Number: num}).Valid
}

func URL(val string) bool {
	_, err := url.ParseRequestURI(val)
	return err == nil
}

func Email(val string) bool {
	_, err := mail.ParseAddress(val)
	return err == nil
}

func IP4(val string) bool {
	ip, err := netip.ParseAddr(val)
	return err == nil && ip.Is4()
}

func IP6(val string) bool {
	ip, err := netip.ParseAddr(val)
	return err == nil && ip.Is6()
}

// IP 判断是否 IP 地址
//
// 同时支持 [IP4] 和 [IP6]
func IP(val string) bool {
	_, err := netip.ParseAddr(val)
	return err == nil
}

// ISBN 判断是否为合法的 [ISBN] 串号
//
// 可以同时判断 [ISBN10] 和 [ISBN13]
//
// [ISBN]: https://zh.wikipedia.org/wiki/%E5%9B%BD%E9%99%85%E6%A0%87%E5%87%86%E4%B9%A6%E5%8F%B7
func ISBN(val string) bool { return isbn.IsValid([]byte(val)) }

func ISBN10(val string) bool { return isbn.ISBN10([]byte(val)) }

func ISBN13(val string) bool { return isbn.ISBN13([]byte(val)) }

// GB11643 判断一个身份证是否符合 gb11643 标准
//
// 若是 15 位则当作一代身份证，仅简单地判断各位是否都是数字；
// 若是 18 位则当作二代身份证，会计算校验位是否正确；
func GB11643(val string) bool { return gb11643.IsValid([]byte(val)) }

// GB32100 统一信用代码校验
func GB32100(val string) bool { return gb32100.IsValid([]byte(val)) }

// BankCard 是否为正确的银行卡号
func BankCard(val string) bool { return Luhn(val) }

// Luhn 验证 [luhn] 算法
//
// [luhn]: https://en.wikipedia.org/wiki/Luhn_algorithm
func Luhn(val string) bool { return luhn.IsValid([]byte(val)) }

// HexColor 判断一个字符串是否为合法的 16 进制颜色表示法
func HexColor(val string) bool {
	if (len(val) != 4 && len(val) != 7) || val[0] != '#' {
		return false
	}

	return Hex(val[1:])
}

// Hex 是否符合 16 进制数字
func Hex(val string) bool {
	for _, v := range val {
		switch {
		case '0' <= v && v <= '9':
		case 'a' <= v && v <= 'f':
		case 'A' <= v && v <= 'F':
		default:
			return false
		}
	}
	return true
}

func StartWith(prefix string) func(string) bool {
	return func(s string) bool { return strings.HasPrefix(s, prefix) }
}

func EndWith(suffix string) func(string) bool {
	return func(s string) bool { return strings.HasSuffix(s, suffix) }
}

// ASCII ASCII 码
func ASCII(s string) bool {
	for _, c := range s {
		if c > 127 {
			return false
		}
	}
	return true
}

// Alpha 全部都是英文字符
func Alpha(s string) bool {
	for _, c := range s {
		if c < 'a' || c > 'z' && (c < 'A' || c > 'Z') {
			return false
		}
	}
	return true
}

// CNMobile 验证中国大陆的手机号码
func CNMobile(val string) bool {
	// 可选的 0,86,086,+86
	// 开头 12x-19x

	if len(val) == 0 {
		return false
	}

	if val[0] == '0' || val[0] == '+' {
		val = val[1:]
	}

	val = strings.TrimPrefix(val, "86")

	if len(val) != 11 {
		return false
	}

	if val[0] != '1' {
		return false
	}

	return Digit(val[1:])
}

// Digit 判断字符串是否都为数字
func Digit(val string) bool {
	for _, c := range val[1:] {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// UUID 验证 UUID 格式是否正确
func UUID(val string) bool { return uuid.Validate(val) == nil }

// Empty 字符串是否为空
func Empty(val string) bool { return val == "" }

// LanguageTag 是否能解析为 [language.Tag] 变量
func LanguageTag(val string) bool {
	_, err := language.Parse(val)
	return err == nil
}
