// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package sanitizer

import (
	"bufio"
	"html"
	"strings"
	"unicode"
)

var nl2brReplacer = strings.NewReplacer(
	"\r\n", "<br />",
	"\n", "<br />",
)

// NL2BR 将换行符转换为 <br />
func NL2BR(v *string) { *v = nl2brReplacer.Replace(*v) }

// BR2NL 将 <br /> 转换为换行符
func BR2NL(v *string) {
	old := *v
	l := len(old)
	var b strings.Builder

LOOP:
	for i := 0; i < l; i++ {
		c := old[i]
		if c == '<' && i+3 <= l { // 最起码还需要 br>
			n1, n2 := old[i+1], old[i+2]
			if (n1 != 'b' && n1 != 'B') || (n2 != 'r' && n2 != 'R') {
				b.WriteByte(c)
				continue
			}

			for j := i + 3; j < l; j++ {
				c := old[j]

				if unicode.IsSpace(rune(c)) {
					continue
				}
				if (c == '/' && j+1 <= l && old[j+1] == '>') || c == '>' {
					if c == '/' {
						j++
					}
					b.WriteByte('\n')
				} else {
					b.WriteString(old[i : j+1])
				}

				i = j
				continue LOOP
			}
		}

		b.WriteByte(c)
	}

	*v = b.String()
}

func EscapeHTML(v *string) { *v = html.EscapeString(*v) }

func UnescapeHTML(v *string) { *v = html.UnescapeString(*v) }

// NL2P 将换行符转换为 <p> 包含的元素
//
// 规则如下：
//
//	l1  => <p>l1</p>
//	l2\nl2 => <p>l1</p><p>l2</p>
func NL2P(v *string) {
	var b strings.Builder

	s := bufio.NewScanner(strings.NewReader(*v))
	s.Split(bufio.ScanLines)
	for s.Scan() {
		b.WriteString("<p>")
		b.Write(s.Bytes())
		b.WriteString("</p>")
	}

	*v = b.String()
}

var p2nlReplacer = strings.NewReplacer(
	"<p>", "",
	"</p>", "\n",
	"<P>", "",
	"</P>", "\n",
)

// P2NL 将 <p> 包含的元素替换为换行符
func P2NL(v *string) { *v = p2nlReplacer.Replace(*v) }

var scriptReplacer = strings.NewReplacer(
	"<script>", "&lt;script&gt;<pre>",
	"</script>", "</pre>&lt;/script&gt;",
)

// EscapeScript 仅转换 HTML 中 script 标签内的内容
func EscapeScript(v *string) { *v = scriptReplacer.Replace(*v) }
