// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package sanitizer

import (
	"testing"

	"github.com/issue9/assert/v4"
)

func TestNL2BR(t *testing.T) {
	a := assert.New(t, false)

	v := "abc\nd"
	NL2BR(&v)
	a.Equal(v, "abc<br />d")

	v = "abc\r\nd\ne"
	NL2BR(&v)
	a.Equal(v, "abc<br />d<br />e")

	v = "abc\r\n\nd\ne"
	NL2BR(&v)
	a.Equal(v, "abc<br /><br />d<br />e")

	v = "ab<br />c\r\nd\ne"
	NL2BR(&v)
	a.Equal(v, "ab<br />c<br />d<br />e")
}

func TestBR2NL(t *testing.T) {
	a := assert.New(t, false)

	v := "abc\nd"
	BR2NL(&v)
	a.Equal(v, "abc\nd")

	v = "abc<br>d"
	BR2NL(&v)
	a.Equal(v, "abc\nd")

	v = "abc<br/>d"
	BR2NL(&v)
	a.Equal(v, "abc\nd")

	v = "abc<BR\t/>d"
	BR2NL(&v)
	a.Equal(v, "abc\nd")

	v = "abc<BR\t  \t>d"
	BR2NL(&v)
	a.Equal(v, "abc\nd")

	v = "abc<BR\t  \t><br />d<br>e"
	BR2NL(&v)
	a.Equal(v, "abc\n\nd\ne")

	v = "<h1>abc<BR\t  \t><br />d<br>e</h1>"
	BR2NL(&v)
	a.Equal(v, "<h1>abc\n\nd\ne</h1>")

	v = "<h1>abc<BR1\t  \t><br />d<br>e</h1>"
	BR2NL(&v)
	a.Equal(v, "<h1>abc<BR1\t  \t>\nd\ne</h1>")
}

func TestNL2P(t *testing.T) {
	a := assert.New(t, false)

	v := "l1"
	NL2P(&v)
	a.Equal(v, "<p>l1</p>")

	v = "l1\r\n"
	NL2P(&v)
	a.Equal(v, "<p>l1</p>")

	v = "\nl1\n"
	NL2P(&v)
	a.Equal(v, "<p></p><p>l1</p>")

	v = "\nl1\n\nl2"
	NL2P(&v)
	a.Equal(v, "<p></p><p>l1</p><p></p><p>l2</p>")
}

func TestP2NL(t *testing.T) {
	a := assert.New(t, false)

	v := "l1"
	P2NL(&v)
	a.Equal(v, "l1")

	v = "<p>l1</P>"
	P2NL(&v)
	a.Equal(v, "l1\n")
}

func TestEscapeScript(t *testing.T) {
	a := assert.New(t, false)

	v := "<p>123</p>"
	EscapeScript(&v)
	a.Equal(v, "<p>123</p>")

	v = "<p>123</p><script>var x = 5</script>"
	EscapeScript(&v)
	a.Equal(v, "<p>123</p>&lt;script&gt;<pre>var x = 5</pre>&lt;/script&gt;")
}
