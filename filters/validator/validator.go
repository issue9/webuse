// SPDX-FileCopyrightText: 2022-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package validator 符合 [web.filter] 的验证器
//
// [web.filter]: https://pkg.go.dev/github.com/issue9/web#Filter
package validator

import (
	"encoding/json"
	"reflect"

	"github.com/issue9/web"
	"github.com/issue9/web/filter"
)

// V 同 [filter.V]
func V[T any](v func(T) bool, msg web.LocaleStringer) filter.Rule[T] { return filter.V(v, msg) }

// SV 同 [filter.SV]
func SV[S ~[]T, T any](v func(T) bool, msg web.LocaleStringer) filter.Rule[S] {
	return filter.SV[S](v, msg)
}

// MV 同 [filter.MV]
func MV[M ~map[K]V, K comparable, V any](v func(V) bool, msg web.LocaleStringer) filter.Rule[M] {
	return filter.MV[M](v, msg)
}

// And 以与的形式串联多个验证器函数
func And[T any](v ...func(T) bool) func(T) bool {
	return func(val T) bool {
		for _, validator := range v {
			if !validator(val) {
				return false
			}
		}
		return true
	}
}

// Or 以或的形式并联多个验证器函数
func Or[T any](v ...func(T) bool) func(T) bool {
	return func(val T) bool {
		for _, validator := range v {
			if validator(val) {
				return true
			}
		}
		return false
	}
}

// Not 验证器的取反
func Not[T any](v func(T) bool) func(T) bool { return func(val T) bool { return !v(val) } }

// Zero 是否为零值
//
// 采用 [reflect.Value.IsZero] 判断。
func Zero[T any](v T) bool { return reflect.ValueOf(v).IsZero() }

// Equal 生成判断值是否等于 v 的验证器
func Equal[T comparable](v T) func(T) bool { return func(t T) bool { return t == v } }

// Nil 是否为 nil
func Nil[T any](v T) bool { return reflect.ValueOf(v).IsNil() }

// JSON 验证是否为正确的 JSON 内容
func JSON(val []byte) bool { return json.Valid(val) }
