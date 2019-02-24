// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package compress

import (
	"testing"

	"github.com/issue9/assert"
)

func TestOptions_build(t *testing.T) {
	a := assert.New(t)

	opt := &Options{
		Funcs: map[string]WriterFunc{"gzip": NewGzip},
		Types: []string{"application/xml", "text/*", "application/json"},
		Size:  1024,
	}
	opt.build()
	a.Equal(opt.prefixTypes, []string{"text/"})
	a.Equal(opt.types, []string{"application/xml", "application/json"})
}

func TestOptions_canComporessed(t *testing.T) {
	a := assert.New(t)

	opt := &Options{}
	opt.build()
	a.False(opt.canCompressed(0, ""))

	opt = &Options{
		Funcs: map[string]WriterFunc{"gzip": NewGzip},
		Types: []string{"text/*", "application/json"},
		Size:  1024,
	}
	opt.build()

	// 长度不够
	a.False(opt.canCompressed(10, ""))

	// 长度够，但是未指定 content-type
	a.False(opt.canCompressed(2046, ""))

	a.True(opt.canCompressed(2046, "text/html;charset=utf-8"))

	a.True(opt.canCompressed(2046, "application/json"))

	a.False(opt.canCompressed(2046, "application/octet"))
}
