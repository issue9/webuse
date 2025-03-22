// SPDX-FileCopyrightText: 2025 caixw
//
// SPDX-License-Identifier: MIT

package systat

import (
	"os"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

// Stats 监视的状态信息
type Stats struct {
	XMLName struct{} `json:"-" yaml:"-" xml:"stats" cbor:"-" toml:"-"`

	OS      *OS       `json:"os" yaml:"os" xml:"os" cbor:"os" toml:"os" comment:"os stats"`                               // 系统级别的状态信息
	Process *Process  `json:"process" yaml:"process" xml:"process" cbor:"process" toml:"process" comment:"process stats"` // 当前进程的状态信息
	Created time.Time `json:"created" yaml:"created" xml:"created" cbor:"created" toml:"created" comment:"created time"`  // 此条记录的创建时间
}

// OS 与系统相关的信息
type OS struct {
	CPU float64 `json:"cpu" yaml:"cpu" xml:"cpu" cbor:"cpu" toml:"cpu" comment:"cpu usage rate"`                                              // CPU 使用百分比
	Mem uint64  `json:"mem" yaml:"mem" xml:"mem" cbor:"mem" toml:"mem" comment:"mem usage rate"`                                              // 内存使用量，以 byte 为单位。
	Net *Net    `json:"net,omitempty" yaml:"net,omitempty" xml:"net,omitempty" cbor:"net,omitempty" toml:"net,omitempty" comment:"net stats"` // 网络相关数据
}

// Process 与进程相关的信息
type Process struct {
	CPU        float64 `json:"cpu" yaml:"cpu" xml:"cpu" cbor:"cpu" toml:"cpu" comment:"cpu usage rate"`            // CPU 使用百分比
	Mem        uint64  `json:"mem" yaml:"mem" xml:"mem" cbor:"mem" toml:"mem" comment:"mem usage rate"`            // 内存使用量，以 byte 为单位。
	Conns      int     `json:"conns" yaml:"conns" xml:"conns" cbor:"conns" toml:"conns" comment:"connects number"` // 连接数量
	Goroutines int     `json:"goroutines,omitempty" yaml:"goroutines,omitempty" xml:"goroutines,omitempty" cbor:"goroutines,omitempty" toml:"goroutines,omitempty" comment:"goroutines number"`
}

// Net 与网络相关的信息
type Net struct {
	Conns int    `json:"conns" yaml:"conns" xml:"conns" cbor:"conns" toml:"conns" comment:"connects number"` // 连接数量
	Sent  uint64 `json:"sent" yaml:"sent" xml:"sent" cbor:"sent" toml:"sent" comment:"sent bytes"`           // 发送数量，以字节为单位。
	Recv  uint64 `json:"recv" yaml:"recv" xml:"recv" cbor:"recv" toml:"recv" comment:"recv bytes"`           // 读取数量，以字节为单位。
}

func calcState(now time.Time) (*Stats, error) {
	os, err := calcOS()
	if err != nil {
		return nil, err
	}

	p, err := calcProcess()
	if err != nil {
		return nil, err
	}

	return &Stats{OS: os, Process: p, Created: now}, nil
}

func calcProcess() (*Process, error) {
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

	return &Process{
		CPU:        cpus,
		Mem:        mems.RSS,
		Conns:      len(conns),
		Goroutines: runtime.NumGoroutine(),
	}, nil
}

func calcOS() (*OS, error) {
	cpus, err := cpu.Percent(0, false)
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

	return &OS{
		CPU: cpus[0],
		Mem: mems.Used,
		Net: &Net{
			Conns: len(conn),
			Sent:  netIO[0].BytesSent,
			Recv:  netIO[0].BytesRecv,
		},
	}, nil
}
