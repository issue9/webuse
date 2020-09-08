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

func TestOptions_canCompressed(t *testing.T) {
	a := assert.New(t)

	opt := &Options{}
	opt.build()
	a.False(opt.canCompressed(""))

	opt = &Options{
		Funcs: map[string]WriterFunc{"gzip": NewGzip},
		Types: []string{"text/*", "application/json"},
	}
	opt.build()

	// 长度不够
	a.False(opt.canCompressed(""))

	// 长度够，但是未指定 content-type
	a.False(opt.canCompressed(""))

	a.True(opt.canCompressed("text/html;charset=utf-8"))

	a.True(opt.canCompressed("application/json"))

	a.False(opt.canCompressed("application/octet"))
}
