// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package validator

import "time"

// Timezone 是否为一个正确的时区变量
func Timezone(tz string) bool {
	_, err := time.LoadLocation(tz)
	return err == nil
}
