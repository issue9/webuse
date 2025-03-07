// SPDX-FileCopyrightText: 2022-2024 caixw
//
// SPDX-License-Identifier: MIT

package isbn

import "bytes"

// IsValid 判断是否为合法的 ISBN10 和 ISBN13 串号
func IsValid(val []byte) bool {
	if bytes.IndexByte(val, '-') > -1 {
		val = eraseMinus(val)
	}

	switch len(val) {
	case 10:
		return isISBN10(val)
	case 13:
		return isISBN13(val)
	default:
		return false
	}
}

// ISBN10 判断是否为合法的 ISBN10
func ISBN10(val []byte) bool {
	if bytes.IndexByte(val, '-') == -1 {
		return isISBN10(val)
	}
	return isISBN10(eraseMinus(val))
}

// ISBN13 判断是否为合法的 ISBN13
func ISBN13(val []byte) bool {
	if bytes.IndexByte(val, '-') == -1 {
		return isISBN13(val)
	}
	return isISBN13(eraseMinus(val))
}

// isbn10 的校验位对应的值
var isbn10Map = []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'x', '0'}

func isISBN10(val []byte) bool {
	sum := 0
	for i := range 9 {
		sum += int(val[i]-'0') * (10 - i)
	}

	if val[9] == 'X' {
		val[9] = 'x'
	}

	return isbn10Map[11-sum%11] == val[9]
}

func isISBN13(val []byte) bool {
	sum := 0
	for i := 0; i < 12; i += 2 {
		sum += int(val[i] - '0')
	}

	for i := 1; i < 12; i += 2 {
		sum += (int(val[i]-'0') * 3)
	}

	return (10 - sum%10) == int(val[12]-'0')
}

// 过滤减号(-)符号
func eraseMinus(val []byte) []byte {
	offset := 0
	for k, v := range val {
		if v == '-' {
			offset++
			continue
		}

		val[k-offset] = v
	}
	return val[:len(val)-offset]
}
