// SPDX-FileCopyrightText: 2022-2025 caixw
//
// SPDX-License-Identifier: MIT

package strength

import (
	"unicode"

	"github.com/issue9/rands/v3"
)

var chars = []byte("abcdefghijkmnprstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ1234567890!@#$%^&*()_+[]{};':\",./<>?")

func lower() []byte    { return chars[0:23] }
func upper() []byte    { return chars[23:47] }
func num() []byte      { return chars[47:57] }
func punct() []byte    { return chars[57:] }
func allChars() []byte { return chars }

// Strength 密码强度管理
type Strength struct {
	Length int8 // 长度不能小于此值
	Upper  int8 // 大写字母的数量不能小于此值
	Lower  int8 // 小写字母的数量不能小于此值
	Punct  int8 // 符号的数量不能小于此值
	Number int8 // 数值的数量不能小于此值
}

// Gen 生成符合要求的随机密码
func (s *Strength) Gen() []byte {
	bs := rands.Bytes(int(s.Length), int(s.Length+1), allChars())
	cnt := Strength{}
	for _, b := range bs {
		switch r := rune(b); {
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			cnt.Punct++
		case unicode.IsUpper(r):
			cnt.Upper++
		case unicode.IsLower(r):
			cnt.Lower++
		case unicode.IsNumber(r):
			cnt.Number++
		}
	}

	if size := s.Punct - cnt.Punct; size > 0 {
		bs = append(bs, rands.Bytes(int(size), int(size+3), punct())...)
	}

	if size := s.Upper - cnt.Upper; size > 0 {
		bs = append(bs, rands.Bytes(int(size), int(size+3), upper())...)
	}

	if size := s.Lower - cnt.Lower; size > 0 {
		bs = append(bs, rands.Bytes(int(size), int(size+3), lower())...)
	}

	if size := s.Number - cnt.Number; size > 0 {
		bs = append(bs, rands.Bytes(int(size), int(size+3), num())...)
	}

	return bs
}

// Valid 验证密码是否符合要求
func (s *Strength) Valid(pass string) bool {
	if s.Length == 0 && s.Upper == 0 && s.Lower == 0 && s.Punct == 0 {
		return true
	}

	cnt := Strength{}
	for _, r := range pass {
		switch {
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			cnt.Punct++
		case unicode.IsUpper(r):
			cnt.Upper++
		case unicode.IsLower(r):
			cnt.Lower++
		case unicode.IsNumber(r):
			cnt.Number++
		}
		cnt.Length++
	}

	ok := true
	if s.Length > 0 {
		ok = cnt.Length >= s.Length
	}
	if ok && s.Lower > 0 {
		ok = cnt.Lower >= s.Lower
	}
	if ok && s.Upper > 0 {
		ok = cnt.Upper >= s.Upper
	}
	if ok && s.Punct > 0 {
		ok = cnt.Punct >= s.Punct
	}
	if ok && s.Number > 0 {
		ok = cnt.Number >= s.Number
	}
	return ok
}
