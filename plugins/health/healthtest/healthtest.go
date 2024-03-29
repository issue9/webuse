// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

// Package healthtest 提供对 [health.Store] 的测试用例
package healthtest

import (
	"time"

	"github.com/issue9/assert/v4"

	"github.com/issue9/webuse/v7/plugins/health"
)

func Test(a *assert.Assertion, s health.Store) {
	state1 := s.Get("r1", "GET", "/users/{id}")
	a.NotNil(state1).
		Equal(state1.Pattern, "/users/{id}").
		Equal(state1.Spend, 0).
		Zero(state1.Count)

	state := &health.State{
		Route:        "r1",
		Method:       "GET",
		Pattern:      "/users/{id}",
		Min:          100,
		Max:          200,
		Count:        5,
		UserErrors:   0,
		ServerErrors: 1,
		Last:         time.Now(),
		Spend:        3000,
	}
	s.Save(state)
	state2 := s.Get("r1", "GET", "/users/{id}")
	a.Equal(state2.Count, 5)

	a.Equal(s.All(), []*health.State{state2})
}
