// SPDX-FileCopyrightText: 2022-2024 caixw
//
// SPDX-License-Identifier: MIT

package validator

import "testing"

func BenchmarkCNMobile(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CNMobile("15011111111")
	}
}
