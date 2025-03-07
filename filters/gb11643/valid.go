// SPDX-FileCopyrightText: 2022-2025 caixw
//
// SPDX-License-Identifier: MIT

package gb11643

var (
	// 校验位对应的规则。
	gb11643Map = []byte{'1', '0', 'x', '9', '8', '7', '6', '5', '4', '3', '2'}

	// 前 17 位号码对应的权值，为一个固定数组。可由 gb11643_test.getWeight() 计算得到。
	gb11643Weight = []int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
)

// IsValid 判断一个身份证是否符合 gb11643 标准
func IsValid(val []byte) bool {
	if len(val) == 15 {
		// 15 位，只检测是否包含非数字字符。
		for i := range 15 {
			if val[i] < '0' || val[i] > '9' {
				return false
			}
		} // end for
		return true
	}

	if len(val) != 18 {
		return false
	}

	sum := 0
	for i := range 17 {
		sum += (gb11643Weight[i] * int((val[i] - '0')))
	}
	if val[17] == 'X' {
		val[17] = 'x'
	}
	return gb11643Map[sum%11] == val[17]
}
