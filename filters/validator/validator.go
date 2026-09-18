// SPDX-FileCopyrightText: 2022-2026 caixw
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
)

// V 同 [web.ValidatorRule]
func V[T any](v func(T) bool, msg web.LocaleStringer) web.Rule[T] { return web.ValidatorRule(v, msg) }

// SV 同 [web.SliceValidatorRule]
func SV[S ~[]T, T any](v func(T) bool, msg web.LocaleStringer) web.Rule[S] {
	return web.SliceValidatorRule[S](v, msg)
}

// MV 同 [web.MapValidatorRule]
func MV[M ~map[K]V, K comparable, V any](v func(V) bool, msg web.LocaleStringer) web.Rule[M] {
	return web.MapValidatorRule[M](v, msg)
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
