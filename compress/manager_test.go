// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package compress

import (
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert"
)

func TestManager_canComporessed(t *testing.T) {
	a := assert.New(t)

	mgr := NewManager(nil, nil, 0)
	w := httptest.NewRecorder()
	a.False(mgr.canCompressed(w, nil))

	mgr = NewManager(map[string]WriterFunc{"gzip": NewGzip}, []string{"text/*", "application/json"}, 1024)
	w = httptest.NewRecorder()

	// 长度不够
	w.Header().Set("Content-Length", "10")
	a.False(mgr.canCompressed(w, nil))

	// 长度够，但是未指定 content-type
	w.Header().Set("Content-Length", "2046")
	a.False(mgr.canCompressed(w, nil))

	w.Header().Set("Content-Type", "text/html;charset=utf-8")
	a.True(mgr.canCompressed(w, nil))

	w.Header().Set("Content-Type", "application/json")
	a.True(mgr.canCompressed(w, nil))

	w.Header().Set("Content-Type", "application/octet")
	a.False(mgr.canCompressed(w, nil))
}
