// SPDX-FileCopyrightText: 2022-2024 caixw
//
// SPDX-License-Identifier: MIT

package validator

import "cmp"

// Between 判断数值区间 (min, max)
func Between[T cmp.Ordered](min, max T) func(T) bool {
	if max < min {
		panic("max 必须大于等于 min")
	}
	return func(val T) bool { return val > min && val < max }
}

// BetweenEqual 判断数值区间 [min, max]
func BetweenEqual[T cmp.Ordered](min, max T) func(T) bool {
	if max < min {
		panic("max 必须大于等于 min")
	}
	return func(val T) bool { return val >= min && val <= max }
}

func Less[T cmp.Ordered](num T) func(T) bool { return func(t T) bool { return t < num } }

func LessEqual[T cmp.Ordered](num T) func(T) bool { return func(t T) bool { return t <= num } }

func Great[T cmp.Ordered](num T) func(T) bool { return func(t T) bool { return t > num } }

func GreatEqual[T cmp.Ordered](num T) func(T) bool { return func(t T) bool { return t >= num } }

// HTTPStatus 是否为有效的 HTTP 状态码
func HTTPStatus(s int) bool { return BetweenEqual(100, 599)(s) }

// ZeroOr 判断值为零值或是非零情况下符合 v 的要求
func ZeroOr[T comparable](v func(T) bool) func(T) bool {
	var zero T
	return Or(func(v T) bool { return v == zero }, v)
}
