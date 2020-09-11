// SPDX-License-Identifier: MIT

package digest

import (
	"testing"
	"time"

	"github.com/issue9/assert"
)

func TestNonces(t *testing.T) {
	a := assert.New(t)
	n, err := newNonces(3*time.Second, 1*time.Second)
	a.NotError(err).NotNil(n)

	n1 := n.newNonce()
	a.NotEmpty(n1.key).
		Equal(n1.count, 0)

	n2 := n.newNonce()
	a.NotEmpty(n2.key).
		Equal(n2.count, 0)
	a.NotEqual(n1.key, n2.key)

	// 测试 GC
	a.Equal(2, len(n.nonces))
	time.Sleep(1 * time.Second)
	n2.setCount(2) // 续命
	time.Sleep(1 * time.Second)
	n2.setCount(4) // 续命
	time.Sleep(1 * time.Second)
	n2.setCount(5) // 续命
	time.Sleep(1 * time.Second)
	n2.setCount(7) // 续命
	a.Equal(1, len(n.nonces)).
		Nil(n.get(n1.key)).
		NotNil(n.get(n2.key))
}

func TestNonce_setCount(t *testing.T) {
	a := assert.New(t)

	n := &nonce{}
	a.NotError(n.setCount(1))
	a.Equal(1, n.count)

	a.NotError(n.setCount(3))
	a.Equal(3, n.count)

	a.Error(n.setCount(1))
	a.Equal(3, n.count)
}
