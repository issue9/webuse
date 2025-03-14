// SPDX-FileCopyrightText: 2025 caixw
//
// SPDX-License-Identifier: MIT

package sanitizer

import "math"

// Abs 转为绝对值
func Abs[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~float32 | ~float64](v *T) {
	*v = T(math.Abs(float64(*v)))
}
