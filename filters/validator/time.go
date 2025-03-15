// SPDX-FileCopyrightText: 2024-2025 caixw
//
// SPDX-License-Identifier: MIT

package validator

import "time"

// Timezone 是否为一个正确的时区变量
func Timezone(tz string) bool {
	_, err := time.LoadLocation(tz)
	return err == nil
}

// Before 判断时间是否在 t 之前
func Before(t time.Time) func(time.Time) bool {
	return func(v time.Time) bool { return v.Before(t) }
}

// After 判断时间是否在 t 之后
func After(t time.Time) func(time.Time) bool {
	return func(v time.Time) bool { return v.After(t) }
}
