// Copyright 2018 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package digest

import (
	"time"

	"github.com/issue9/rands"
	"github.com/issue9/unique"
)

// 服务端随机字符串的管理工具。
type nonces struct {
	nonces  map[string]*nonce
	rands   *unique.Rands
	expired time.Duration
}

type nonce struct {
	key   string    // 随机字符串的值
	count int       // 计数
	last  time.Time // 最后更新时间，超过一定时间未用，会被收回
}

func newNonces(expired, gc time.Duration) (*nonces, error) {
	seed := time.Now().Unix()

	rands, err := rands.New(seed, 1000, 32, 33, rands.AlphaNumber)
	if err != nil {
		return nil, err
	}

	ns := &nonces{
		nonces: make(map[string]*nonce, 1000),
		rands: &unique.Rands{
			Rands:  rands,
			Unique: unique.New(seed, 1, 120, "", 35),
		},
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
	expired := now.Add(-n.expired)

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
