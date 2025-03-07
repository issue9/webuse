// SPDX-FileCopyrightText: 2022-2024 caixw
//
// SPDX-License-Identifier: MIT

package validator

import (
	"testing"

	"github.com/issue9/assert/v4"
)

func TestHexColor(t *testing.T) {
	a := assert.New(t, false)

	a.True(HexColor("#123")).
		True(HexColor("#fff")).
		True(HexColor("#f0f0f0")).
		True(HexColor("#fafafa")).
		True(HexColor("#224422"))

	a.False(HexColor("#2244")).
		False(HexColor("#hhh")).
		False(HexColor("ffff")).
		False(HexColor("#asdf")).
		False(HexColor("#ffff"))
}

func TestIP(t *testing.T) {
	a := assert.New(t, false)

	a.True(IP("fe80:0000:0000:0000:0204:61ff:fe9d:f156")).
		True(IP("fe80:0:0:0:204:61ff:fe9d:f156")).
		True(IP("0.0.0.0")).
		True(IP("255.255.255.255")).
		True(IP("255.0.3.255"))

	a.False(IP("255.0:3.255")).
		False(IP("275.0.3.255"))
}

func TestIP6(t *testing.T) {
	a := assert.New(t, false)

	a.True(IP6("fe80:0000:0000:0000:0204:61ff:fe9d:f156"))      // full form of IPv6
	a.True(IP6("fe80:0:0:0:204:61ff:fe9d:f156"))                // drop leading zeroes
	a.True(IP6("fe80::204:61ff:fe9d:f156"))                     // collapse multiple zeroes to :: in the IPv6 address
	a.True(IP6("fe80:0000:0000:0000:0204:61ff:254.157.241.86")) // IPv4 dotted quad at the end
	a.True(IP6("fe80:0:0:0:0204:61ff:254.157.241.86"))          // drop leading zeroes, IPv4 dotted quad at the end
	a.True(IP6("fe80::204:61ff:254.157.241.86"))                // dotted quad at the end, multiple zeroes collapsed
	a.True(IP6("::1"))                                          // localhost
	a.True(IP6("fe80::"))                                       // link-local prefix
	a.True(IP6("2001::"))                                       // global unicast prefix
}

func TestIP4(t *testing.T) {
	a := assert.New(t, false)

	a.True(IP4("0.0.0.0")).
		True(IP4("255.255.255.255")).
		True(IP4("255.0.3.255")).
		True(IP4("127.10.0.1")).
		True(IP4("27.1.0.1"))

	a.False(IP4("1127.01.0.1"))
}

func TestEmail(t *testing.T) {
	a := assert.New(t, false)

	a.True(Email("email@email.com")).
		True(Email("em2il@email.com.cn")).
		True(Email("12345@qq.com")).
		True(Email("email.test@email.com")).
		True(Email("email.test@email123.com")).
		True(Email("em2il@email"))

	// 2个@
	a.False(Email("em@2l@email.com"))
	// 没有@
	a.False(Email("email2email.com.cn"))
}

func TestURL(t *testing.T) {
	a := assert.New(t, false)

	a.True(URL("http://www.example.com")).
		True(URL("http://www.example.com/path/?a=b")).
		True(URL("https://www.example.com:88/path1/path2")).
		True(URL("ftp://pwd:user@www.example.com/index.go?a=b")).
		True(URL("pwd:user@www.example.com/path/")).
		True(URL("https://[fe80:0:0:0:204:61ff:fe9d:f156]/path/")).
		True(URL("https://127.0.0.1/path//index.go?arg1=val1&arg2=val/2")).
		True(URL("https://[::1]/path/index.go?arg1=val1"))
}

func TestASCII(t *testing.T) {
	a := assert.New(t, false)

	a.True(ASCII("abc")).
		False(ASCII("\u1000"))
}

func TestAlpha(t *testing.T) {
	a := assert.New(t, false)

	a.False(Alpha("12345")).
		True(Alpha("abc")).
		False(Alpha("abc12"))
}

func TestCNMobile(t *testing.T) {
	a := assert.New(t, false)

	a.True(CNMobile("15011111111")).
		True(CNMobile("015011111111")).
		True(CNMobile("8615011111111")).
		True(CNMobile("+8615011111111")).
		True(CNMobile("+8619911111111"))

	a.False(CNMobile("+86150111111112")). // 尾部多个2
						False(CNMobile("50111111112")).    // 开头少1
						False(CNMobile("+8650111111112")). // 开头少1
						False(CNMobile("8650111111112")).  // 开头少1
						False(CNMobile("154111111112")).   // 不存在的前缀154
						True(CNMobile("15411111111"))
}

func TestLanguageTag(t *testing.T) {
	a := assert.New(t, false)

	a.True(LanguageTag("zh-CN")).False(LanguageTag("xxxx"))
}
