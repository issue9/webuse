// SPDX-FileCopyrightText: 2022-2024 caixw
//
// SPDX-License-Identifier: MIT

package gb11643

import (
	"testing"
	"time"

	"github.com/issue9/assert/v4"
)

func TestParse(t *testing.T) {
	a := assert.New(t, false)

	g, err := Parse("513330199111066159")
	a.NotError(err).NotNil(g)
	date, err := time.Parse(layout, "19911106")
	a.NotError(err)
	a.Equal(g.Raw, "513330199111066159").
		Equal(g.Region, "513330").
		Equal(g.Date, date).
		True(g.IsMale)

	g, err = Parse("33050219880702447X")
	a.NotError(err).NotNil(g)
	date, err = time.Parse(layout, "19880702")
	a.NotError(err)
	a.Equal(g.Raw, "33050219880702447X").
		Equal(g.Region, "330502").
		Equal(g.Date, date).
		True(g.IsMale)

	g, err = Parse("330502880702447")
	a.NotError(err).NotNil(g)
	date, err = time.Parse(layout, "19880702")
	a.NotError(err)
	a.Equal(g.Raw, "330502880702447").
		Equal(g.Region, "330502").
		Equal(g.Date, date).
		True(g.IsMale)

	g, err = Parse("350303900203307")
	a.NotError(err).NotNil(g)
	date, err = time.Parse(layout, "19900203")
	a.NotError(err)
	a.Equal(g.Raw, "350303900203307").
		Equal(g.Region, "350303").
		Equal(g.Date, date).
		True(g.IsMale)

	g, err = Parse("")
	a.Error(err).Nil(g)
}
