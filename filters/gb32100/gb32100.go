// SPDX-FileCopyrightText: 2022-2024 caixw
//
// SPDX-License-Identifier: MIT

// Package gb32100 GB32100-2015 统一信用代码校验
package gb32100

import (
	"maps"

	"github.com/issue9/web/locales"
)

var (
	ministries = map[byte]string{
		'1': "机构编制",
		'2': "外交",
		'3': "司法行政",
		'4': "文化",
		'5': "民政",
		'6': "旅游",
		'7': "宗教",
		'8': "工会",
		'9': "工商",
		'A': "中央军委改革和编制办公室",
		'N': "农业",
		'Y': "其他",
	}

	types = map[byte]map[byte]string{
		'1': {
			'1': "机关",
			'2': "事业单位",
			'3': "中央编办直接管理机构编制的群众团体",
			'9': "其他",
		},
		'2': {
			'1': "外国常驻新闻机构",
			'9': "其他",
		},
		'3': {
			'1': "律师执业机构",
			'2': "公证处",
			'3': "基层法律服务所",
			'4': "司法鉴定机构",
			'5': "仲裁委员会",
			'9': "其他",
		},
		'4': {
			'1': "外国在华文化中心",
			'9': "其他",
		},
		'5': {
			'1': "社会团体",
			'2': "民办非企业单位",
			'3': "基金会",
			'9': "其他",
		},
		'6': {
			'1': "外国旅游部门常驻代表机构",
			'2': "港澳台地区旅游部门常驻内地（大陆）代表机构",
			'9': "其他",
		},
		'7': {
			'1': "宗教活动场所",
			'2': "宗教院校",
			'9': "其他",
		},
		'8': {
			'1': "基层工会",
			'9': "其他",
		},
		'9': {
			'1': "企业",
			'2': "个体工商户",
			'3': "农民专业合作社",
		},
		'A': {
			'1': "军队事业单位",
			'9': "其他",
		},
		'N': {
			'1': "组级集体经济组织",
			'2': "村级集体经济组织",
			'3': "乡镇级集体经济组织",
			'9': "其他",
		},
		'Y': {
			'1': "其它",
		},
	}
)

// GB32100 统一信用代码
type GB32100 struct {
	Raw      string
	Ministry byte   // 登记管理部门
	Type     byte   // 类别
	Region   string // 区域信息，可参考 https://github.com/issue9/cnregion
	ID       string // 主体代码
}

// Parse 解析统一信用代码至 [GB32100]
func Parse(bs string) (*GB32100, error) {
	if !IsValid([]byte(bs)) {
		return nil, locales.ErrInvalidFormat()
	}

	return &GB32100{
		Raw:      bs,
		Ministry: bs[0],
		Type:     bs[1],
		Region:   bs[2:8],
		ID:       bs[8:17],
	}, nil
}

// MinistryName 返回登记管理部门
func (g *GB32100) MinistryName() string { return ministries[g.Ministry] }

// TypeName 返回登记管理部门下的类型名称
func (g *GB32100) TypeName() string { return types[g.Ministry][g.Type] }

// Ministries 所有可用的登记管理部分
func Ministries() map[byte]string { return maps.Clone(ministries) }

// MinistryTypes 指定管理部分下的可用分类信息
func MinistryTypes(ministry byte) map[byte]string { return maps.Clone(types[ministry]) }
