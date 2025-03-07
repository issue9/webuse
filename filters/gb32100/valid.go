// SPDX-FileCopyrightText: 2022-2025 caixw
//
// SPDX-License-Identifier: MIT

package gb32100

var (
	// 用到的编码字符
	codeIndexes = map[byte]int{'0': 0, '1': 1, '2': 2, '3': 3, '4': 4, '5': 5, '6': 6, '7': 7, '8': 8, '9': 9,
		'A': 10, 'B': 11, 'C': 12, 'D': 13, 'E': 14, 'F': 15, 'G': 16, 'H': 17, 'J': 18, 'K': 19, 'L': 20,
		'M': 21, 'N': 22, 'P': 23, 'Q': 24, 'R': 25, 'T': 26, 'U': 27, 'W': 28, 'X': 29, 'Y': 30}

	// 加权因子
	factors = []int{1, 3, 9, 27, 19, 26, 16, 17, 20, 29, 25, 13, 8, 24, 10, 30, 28}
)

// IsValid 验证是否有效
//
// 并不是所有的统一信用代码都能通过验证，部分早期可能会失败，https://v2ex.com/t/573549
func IsValid(bs []byte) bool {
	if len(bs) != 18 {
		return false
	}

	ts, found := types[bs[0]]
	if !found {
		return false
	}
	if _, found = ts[bs[1]]; !found {
		return false
	}

	t := 0
	for i := range 17 {
		t += codeIndexes[bs[i]] * factors[i]
	}
	t = 31 - (t % 31)
	return t == codeIndexes[bs[17]]
}
