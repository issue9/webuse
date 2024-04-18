// SPDX-FileCopyrightText: 2024 caixw
//
// SPDX-License-Identifier: MIT

package monitor

import (
	"os"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

type Stats struct {
	XMLName struct{}  `json:"-" yaml:"-" xml:"stats"`
	OS      *Info     `json:"os" yaml:"os" xml:"os"`                // 系统级别的状态信息
	Process *Info     `json:"process" yaml:"process" xml:"process"` // 当前进程的状态信息
	Created time.Time `json:"created" yaml:"created" xml:"created"` // 此条记录的创建时间
}

type Info struct {
	CPU float64 `json:"cpu" yaml:"cpu" xml:"cpu"` // CPU 使用百分比
	Mem uint64  `json:"mem" yaml:"mem" xml:"mem"` // 内存使用量，以 byte 为单位。

	// 网络相关数据
	Net *Net `json:"net,omitempty" yaml:"net,omitempty" xml:"net,omitempty"`

	// 在全局模式之下为空
	Goroutines int `json:"goroutines,omitempty" yaml:"goroutines,omitempty" xml:"goroutines,omitempty"`
}

type Net struct {
	Conn int    `json:"conn" yaml:"conn" xml:"conn"` // 连接数量
	Sent uint64 `json:"sent" yaml:"sent" xml:"sent"` // 发送数量，以字节为单位。
	Recv uint64 `json:"recv" yaml:"recv" xml:"recv"` // 读取数量，以字节为单位。
}

func calcState(interval time.Duration, now time.Time) (*Stats, error) {
	all, err := calcOS(interval)
	if err != nil {
		return nil, err
	}

	p, err := calcProcess()
	if err != nil {
		return nil, err
	}

	return &Stats{OS: all, Process: p, Created: now}, nil
}

func calcProcess() (*Info, error) {
	p, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		return nil, err
	}

	cpus, err := p.CPUPercent()
	if err != nil {
		return nil, err
	}

	mems, err := p.MemoryInfo()
	if err != nil {
		return nil, err
	}

	conns, err := p.Connections()
	if err != nil {
		return nil, err
	}

	netIO, err := p.IOCounters()
	if err != nil {
		return nil, err
	}

	return &Info{
		CPU: cpus,
		Mem: mems.RSS,
		Net: &Net{
			Conn: len(conns),
			Sent: netIO.WriteBytes,
			Recv: netIO.ReadBytes,
		},
		Goroutines: runtime.NumGoroutine(),
	}, nil
}

func calcOS(interval time.Duration) (*Info, error) {
	cpus, err := cpu.Percent(interval, false)
	if err != nil {
		return nil, err
	}

	mems, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	netIO, err := net.IOCounters(false)
	if err != nil {
		return nil, err
	}

	conn, err := net.Connections("all")
	if err != nil {
		return nil, err
	}

	return &Info{
		CPU: cpus[0],
		Mem: mems.Used,
		Net: &Net{
			Conn: len(conn),
			Sent: netIO[0].BytesSent,
			Recv: netIO[0].BytesRecv,
		},
	}, nil
}
