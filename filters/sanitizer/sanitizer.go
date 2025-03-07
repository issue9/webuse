// SPDX-FileCopyrightText: 2022-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package sanitizer 内容修正工具
package sanitizer

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
	"unicode"

	"github.com/issue9/web/filter"
)

// S 同 [filter.S]
func S[T any](f ...func(*T)) filter.Rule[T] { return filter.S(f...) }

// SS 同 [filter.SS]
func SS[S ~[]T, T any](f ...func(*T)) filter.Rule[S] { return filter.SS[S](f...) }

// MS 同 [filter.MS]
func MS[M ~map[K]V, K comparable, V any](f func(*V)) filter.Rule[M] { return filter.MS[M](f) }

// Sanitizers 将多个修正函数合并为一个
func Sanitizers[T any](f ...func(*T)) func(*T) {
	return func(v *T) {
		for _, ss := range f {
			ss(v)
		}
	}
}

// Trim 过滤左右空格
func Trim(v *string) { *v = strings.TrimSpace(*v) }

func TrimLeft(v *string) {
	*v = strings.TrimLeftFunc(*v, func(r rune) bool { return unicode.IsSpace(r) })
}

func TrimRight(v *string) {
	*v = strings.TrimRightFunc(*v, func(r rune) bool { return unicode.IsSpace(r) })
}

func Upper(v *string) { *v = strings.ToUpper(*v) }

func Lower(v *string) { *v = strings.ToLower(*v) }

func MD5(v *string) {
	h := md5.New()
	h.Write([]byte(*v))
	*v = hex.EncodeToString(h.Sum(nil))
}
