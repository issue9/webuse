// SPDX-FileCopyrightText: 2022-2024 caixw
//
// SPDX-License-Identifier: MIT

package gb32100

import (
	"testing"

	"github.com/issue9/assert/v4"
)

var validData = []string{
	"91350100M000100Y43", // 官方示例
	"91310000MA1K35Y38P", // 锤子
	"91110108795101314G", // 谷歌
	"914403001922038216", // 华为
}

func TestParse(t *testing.T) {
	a := assert.New(t, false)

	g, err := Parse(validData[0])
	a.NotError(err).NotNil(g)
	a.Equal(g.MinistryName(), "工商").
		Equal(g.TypeName(), "企业").
		Equal(g.Region, "350100").
		Equal(g.ID, "M000100Y4").
		Equal(g.Raw, validData[0])
}
