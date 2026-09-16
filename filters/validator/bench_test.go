// SPDX-FileCopyrightText: 2022-2026 caixw
//
// SPDX-License-Identifier: MIT

package validator

import "testing"

func BenchmarkCNMobile(b *testing.B) {
	for b.Loop() {
		CNMobile("15011111111")
	}
}
