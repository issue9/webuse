// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package accept 用于处理 accpet 系列的报头。
//
// Deprecated: 已不在使用，请使用 qheader 包的相关内容
package accept

import "github.com/issue9/qheader"

// Accept 表示 Accept* 的报头元素
type Accept = qheader.Header

// Parse 将报头内容解析为 []*Accept，并对内容进行排序之后返回。
//
//
// 排序方式如下:
//
// Q 值大的靠前，如果 Q 值相同，则全名的比带通配符的靠前，*/* 最后。
//
// q 值为 0 的数据将被过滤，比如：
//  application/*;q=0.1,application/xml;q=0.1,text/html;q=0
// 其中的 text/html 不会被返回，application/xml 的优先级会高于 applicatioon/*
func Parse(header string) ([]*Accept, error) {
	return qheader.Parse(header)
}
