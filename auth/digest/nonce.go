// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package digest

import (
	"time"

	"github.com/issue9/rands"
)

var randBytes = []byte("01234567890abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz")

// 服务端随机字符串的管理工具。
type nonces struct {
	nonces  map[string]*nonce
	rands   *rands.Rands
	expired time.Duration
}

type nonce struct {
	key   string
	count int
	last  time.Time
}

func newNonces(expired, gc time.Duration) (*nonces, error) {
	rands, err := rands.New(time.Now().Unix(), 1000, 32, 33, randBytes)
	if err != nil {
		return nil, err
	}

	ns := &nonces{
		nonces:  make(map[string]*nonce, 1000),
		rands:   rands,
		expired: expired,
	}

	go func() {
		for now := range time.Tick(gc) {
			ns.gc(now)
		}
	}()

	return ns, nil
}

func (n *nonces) gc(now time.Time) {
	expired := now.Truncate(n.expired)

	for k, v := range n.nonces {
		if v.last.Before(expired) {
			delete(n.nonces, k)
		}
	}
}

func (n *nonces) get(nonceKey string) *nonce {
	return n.nonces[nonceKey]
}

func (n *nonces) add(nonceKey string) {
	v, found := n.nonces[nonceKey]
	if !found {
		v = &nonce{key: nonceKey}
		n.nonces[nonceKey] = v
	}

	v.count++
	v.last = time.Now()
}

func (n *nonces) newNonce() *nonce {
	nn := &nonce{
		key:  n.rands.String(),
		last: time.Now(),
	}
	n.nonces[nn.key] = nn

	return nn
}
