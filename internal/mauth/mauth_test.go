// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package mauth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/mux/v8/types"
	"github.com/issue9/web/server"
)

func TestGetSet(t *testing.T) {
	a := assert.New(t, false)

	s, err := server.NewHTTP("test", "1.0.0", nil)
	a.NotError(err).NotNil(s)

	ctx := s.NewContext(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/path", nil), types.NewContext())
	val, found := Get[int](ctx)
	a.False(found).Zero(val)

	Set(ctx, 5)
	val, found = Get[int](ctx)
	a.True(found).Equal(val, 5)
}
