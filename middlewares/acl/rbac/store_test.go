// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package rbac_test

import (
	"testing"

	"github.com/issue9/assert/v4"

	"github.com/issue9/webuse/v7/internal/testserver"
	"github.com/issue9/webuse/v7/middlewares/acl/rbac"
	"github.com/issue9/webuse/v7/middlewares/acl/rbac/rbactest"
)

func TestCacheStore(t *testing.T) {
	a := assert.New(t, false)
	s := testserver.New(a)

	rbactest.Test(a, rbac.NewCacheStore[int64](s, "int64_"))
	rbactest.Test(a, rbac.NewCacheStore[string](s, "string_"))
}
