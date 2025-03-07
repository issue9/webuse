// SPDX-FileCopyrightText: 2022-2024 caixw
//
// SPDX-License-Identifier: MIT

package gb11643

import (
	"math"
	"testing"

	"github.com/issue9/assert/v4"
)

// 计算各个数值位对应的系数值。
func getWeight() []int {
	l := 17
	ret := make([]int, 17)
	for i := 0; i < l; i++ {
		k := int(math.Pow(2, float64((l - i)))) // k值足够大，不能用byte保存
		ret[i] = k % 11
	}
	return ret
}

func TestGetWeight(t *testing.T) {
	assert.New(t, false).Equal(gb11643Weight, getWeight())
}

func TestIsValid(t *testing.T) {
	a := assert.New(t, false)

	// 网上扒来的身份证，不与现实中的对应。
	a.True(IsValid([]byte("350303199002033073")))
	a.True(IsValid([]byte("350303900203307")))
	a.True(IsValid([]byte("331122197905116239")))
	a.True(IsValid([]byte("513330199111066159")))
	a.True(IsValid([]byte("33050219880702447x")))
	a.True(IsValid([]byte("33050219880702447X")))
	a.True(IsValid([]byte("330502880702447")))

	a.False(IsValid([]byte("111111111111111111")))
	a.False(IsValid([]byte("330502198807024471"))) // 最后一位不正确
	a.False(IsValid([]byte("33050288070244")))     // 14 位
	a.False(IsValid([]byte("3305028807024411")))   // 16 位
	a.False(IsValid([]byte("33050288070244y")))    // 非法字符
}
