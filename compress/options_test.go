// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package compress

import (
	"net/http/httptest"
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
	w := httptest.NewRecorder()
	a.False(opt.canCompressed(w))

	opt = &Options{
		Funcs: map[string]WriterFunc{"gzip": NewGzip},
		Types: []string{"text/*", "application/json"},
		Size:  1024,
	}
	opt.build()
	w = httptest.NewRecorder()

	// 长度不够
	w.Header().Set("Content-Length", "10")
	a.False(opt.canCompressed(w))

	// 长度够，但是未指定 content-type
	w.Header().Set("Content-Length", "2046")
	a.False(opt.canCompressed(w))

	w.Header().Set("Content-Type", "text/html;charset=utf-8")
	a.True(opt.canCompressed(w))

	w.Header().Set("Content-Type", "application/json")
	a.True(opt.canCompressed(w))

	w.Header().Set("Content-Type", "application/octet")
	a.False(opt.canCompressed(w))
}
