// SPDX-FileCopyrightText: 2024-2025 caixw
//
// SPDX-License-Identifier: MIT

package mauth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/issue9/assert/v4"
	"github.com/issue9/mux/v9/types"

	"github.com/issue9/webuse/v7/internal/testserver"
)

func TestGetSet(t *testing.T) {
	a := assert.New(t, false)
	s := testserver.New(a)

	ctx := s.NewContext(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/path", nil), types.NewContext())
	val, found := Get[int](ctx)
	a.False(found).Zero(val)

	Set(ctx, 5)
	val, found = Get[int](ctx)
	a.True(found).Equal(val, 5)
}
