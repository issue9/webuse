// SPDX-FileCopyrightText: 2022-2026 caixw
//
// SPDX-License-Identifier: MIT

// Package sanitizer 内容修正工具
package sanitizer

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
	"unicode"

	"github.com/issue9/web"
)

// S 同 [web.SanitizeRule]
func S[T any](f ...func(*T)) web.Rule[T] { return web.SanitizerRule(f...) }

// SS 同 [web.SliceSanitizeRule]
func SS[S ~[]T, T any](f ...func(*T)) web.Rule[S] { return web.SliceSanitizerRule[S](f...) }

// MS 同 [web.MapSanitizeRule]
func MS[M ~map[K]V, K comparable, V any](f func(*V)) web.Rule[M] { return web.MapSanitizerRule[M](f) }

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
