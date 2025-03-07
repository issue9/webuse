// SPDX-FileCopyrightText: 2022-2024 caixw
//
// SPDX-License-Identifier: MIT

package validator

import "regexp"

// 匹配大陆电话
const cnPhonePattern = `((\d{3,4})-?)?` + // 区号
	`\d{5,10}` + // 号码，95500 等 5 位数的，7 位，8 位，以及 400 开头的 10 位数
	`(-\d{1,4})?` // 分机号，分机号的连接符号不能省略。

var cnPhone = regexpCompile(cnPhonePattern)

func regexpCompile(str string) *regexp.Regexp { return regexp.MustCompile("^" + str + "$") }

// CNPhone 验证中国大陆的电话号码
//
// 支持如下格式：
//
//	0578-12345678-1234
//	057812345678-1234
//
// 若存在分机号，则分机号的连接符不能省略。
func CNPhone(val string) bool { return Match(cnPhone)(val) }

// CNTel 验证手机和电话类型
func CNTel(val string) bool { return CNMobile(val) || CNPhone(val) }

// Match 为正则生成验证函数
func Match(exp *regexp.Regexp) func(string) bool {
	return func(val string) bool { return exp.MatchString(val) }
}

// Regexp 是否为一个正确的正则表达式
func Regexp(v string) bool {
	_, err := regexp.Compile(v)
	return err == nil
}
