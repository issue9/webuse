// SPDX-FileCopyrightText: 2022-2024 caixw
//
// SPDX-License-Identifier: MIT

package validator

import "slices"

// In 声明枚举类型的验证规则
//
// 要求验证的值必须包含在 element 元素中，如果不存在，则返回 msg 的内容。
func In[T comparable](element ...T) func(T) bool {
	return func(v T) bool { return slices.Index(element, v) >= 0 }
}

// NotIn 声明不在枚举中的验证规则
func NotIn[T comparable](element ...T) func(T) bool { return Not(In(element...)) }
