package main

import (
	"fmt"
	"runtime"
	"sort"
	"time"

	"go-tools/tools/mysql"
	"go-tools/tools/redis"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

// SystemInfo 系统概览信息（供前端面板展示）
type SystemInfo struct {
	Hostname    string      `json:"hostname"`
	OS          string      `json:"os"`
	Platform    string      `json:"platform"`
	Uptime      uint64      `json:"uptime"`
	CPUUsage    float64     `json:"cpuUsage"`
	CPUCores    int         `json:"cpuCores"`
	MemTotal    uint64      `json:"memTotal"`
	MemUsed     uint64      `json:"memUsed"`
	MemPercent  float64     `json:"memPercent"`
	Goroutines  int         `json:"goroutines"`
	AppMemAlloc uint64      `json:"appMemAlloc"`
	Listeners   []PortEntry `json:"listeners"`
	Services    []SvcEntry  `json:"services"`
}

// PortEntry 一个监听端口
type PortEntry struct {
	Proto   string `json:"proto"`
	Port    int    `json:"port"`
	PID     int    `json:"pid"`
	Process string `json:"process"`
	Address string `json:"address"`
}

// SvcEntry 由本程序管理的本地服务实例状态
type SvcEntry struct {
	ID      string `json:"id"`
	Type    string `json:"type"` // redis | mysql
	Name    string `json:"name"`
	Port    int    `json:"port"`
	Running bool   `json:"running"`
	Version string `json:"version"`
}

// GetSystemInfo 采集系统快照
func (a *App) GetSystemInfo() (*SystemInfo, error) {
	info := &SystemInfo{}

	// 主机
	if h, err := host.Info(); err == nil {
		info.Hostname = h.Hostname
		info.OS = fmt.Sprintf("%s %s", h.OS, h.PlatformVersion)
		info.Platform = h.Platform
		info.Uptime = h.Uptime
	}

	// CPU
	info.CPUCores = runtime.NumCPU()
	if percentages, err := cpu.Percent(200*time.Millisecond, false); err == nil && len(percentages) > 0 {
		info.CPUUsage = percentages[0]
	}

	// 内存
	if vm, err := mem.VirtualMemory(); err == nil {
		info.MemTotal = vm.Total
		info.MemUsed = vm.Used
		info.MemPercent = vm.UsedPercent
	}

	// Go 运行时
	info.Goroutines = runtime.NumGoroutine()
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	info.AppMemAlloc = ms.Alloc

	// 监听端口
	conns, err := net.Connections("tcp")
	if err == nil {
		seen := map[uint32]bool{}
		for _, c := range conns {
			if c.Status != "LISTEN" {
				continue
			}
			key := c.Laddr.Port*10 + uint32(c.Pid)
			if seen[key] {
				continue
			}
			seen[key] = true
			proc := ""
			if c.Pid > 0 {
				if p, err := os_findProcess(c.Pid); err == nil {
					proc = p
				}
			}
			addr := fmt.Sprintf("%s:%d", c.Laddr.IP, c.Laddr.Port)
			info.Listeners = append(info.Listeners, PortEntry{
				Proto: "tcp", Port: int(c.Laddr.Port), PID: int(c.Pid), Process: proc, Address: addr,
			})
		}
		sort.Slice(info.Listeners, func(i, j int) bool {
			return info.Listeners[i].Port < info.Listeners[j].Port
		})
	}

	// 本程序管理的服务实例
	info.Services = collectServiceStatus()

	return info, nil
}

// os_findProcess 获取 PID 对应的进程名
func os_findProcess(pid int32) (string, error) {
	p, err := process.NewProcess(pid)
	if err != nil {
		return "", err
	}
	return p.Name()
}

func collectServiceStatus() []SvcEntry {
	var out []SvcEntry
	// Redis instances
	if redisInsts, _ := redis.ListInstances(); redisInsts != nil {
		for _, inst := range redisInsts {
			running := false
			if st, err := redis.GetInstanceStatus(inst.ID); err == nil {
				running = st.Running
			}
			out = append(out, SvcEntry{ID: inst.ID, Type: "redis", Name: inst.Name, Port: inst.Port, Running: running, Version: inst.Version})
		}
	}
	// MySQL instances
	if mysqlInsts, _ := mysql.ListInstances(); mysqlInsts != nil {
		for _, inst := range mysqlInsts {
			running := false
			if st, err := mysql.GetInstanceStatus(inst.ID); err == nil {
				running = st.Running
			}
			out = append(out, SvcEntry{ID: inst.ID, Type: "mysql", Name: inst.Name, Port: inst.Port, Running: running, Version: inst.Version})
		}
	}
	return out
}

// 过滤掉 ephemeral 端口范围（Windows 默认 49152-65535）
var _ = sort.Ints // keep sort import
