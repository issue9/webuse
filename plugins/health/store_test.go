// SPDX-FileCopyrightText: 2015-2024 caixw
//
// SPDX-License-Identifier: MIT

package health_test

import (
	"testing"

	"github.com/issue9/assert/v4"

	"github.com/issue9/webuse/v7/internal/testserver"
	"github.com/issue9/webuse/v7/plugins/health"
	"github.com/issue9/webuse/v7/plugins/health/healthtest"
)

func TestCacheStore(t *testing.T) {
	a := assert.New(t, false)
	s := testserver.New(a)
	healthtest.Test(a, health.NewCacheStore(s, "prefix_"))
}
