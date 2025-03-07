// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package validator

import "github.com/issue9/version"

// Semver [semver] 版本号验证
//
// [semver]: https://semver.org/lang/zh-CN/
func Semver(ver string) bool { return version.SemVerValid(ver) }

// SemverCompatible 创建一个验证 [semver] 是否兼容 val 的验证器
func SemverCompatible(ver string) func(string) bool {
	semver, err := version.SemVer(ver)
	if err != nil {
		panic(err)
	}

	return func(v string) bool {
		ok, err := semver.CompatibleString(v)
		return err == nil && ok
	}
}

// SemverGreat 判断版本号是否大于 ver
func SemverGreat(ver string) func(string) bool {
	semver, err := version.SemVer(ver)
	if err != nil {
		panic(err)
	}

	return func(val string) bool {
		num, err := semver.CompareString(val)
		return err == nil && num < 0
	}
}

func SemverGreatEqual(ver string) func(string) bool {
	semver, err := version.SemVer(ver)
	if err != nil {
		panic(err)
	}

	return func(val string) bool {
		num, err := semver.CompareString(val)
		return err == nil && num <= 0
	}
}

func SemverLess(ver string) func(string) bool {
	semver, err := version.SemVer(ver)
	if err != nil {
		panic(err)
	}

	return func(val string) bool {
		num, err := semver.CompareString(val)
		return err == nil && num > 0
	}
}

func SemverLessEqual(ver string) func(string) bool {
	semver, err := version.SemVer(ver)
	if err != nil {
		panic(err)
	}

	return func(val string) bool {
		num, err := semver.CompareString(val)
		return err == nil && num >= 0
	}
}
