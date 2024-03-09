// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

// Package compress 根据 CPU 占用情况决定是否启用压缩
package compress

import (
	"time"

	"github.com/issue9/web"
	"github.com/shirou/gopsutil/v3/cpu"
)

// New 根据 CPU 使用率决定是否启用压缩功能
//
// dur 多少时间检测一次 CPU 的使用率，不能小于 [time.Second]；
// interval 每次检测时读取的时间长度，不能大于 dur；
// percent CPU 的使用率大于此值时将禁用压缩功能；
func New(dur, interval time.Duration, percent float64) web.Plugin {
	if dur < interval {
		panic("dur 必须大于 interval")
	}

	if percent < 0 || percent > 100 {
		panic("percent 必须介于 [0,100]")
	}

	return web.PluginFunc(func(s web.Server) {
		s.Services().AddTicker(web.Phrase("enable compression base on cpu used"), func(now time.Time) error {
			if vals, err := cpu.Percent(interval, false); err != nil {
				s.Logs().ERROR().Error(err)
			} else {
				s.SetCompress(vals[0] < percent)
			}

			return nil
		}, dur, true, false)
	})
}
