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
}
