package system

import (
	"context"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"

	v1 "opskvm/api/system/v1"
)

type SysAllLogic struct{}

func NewSysAllLogic() *SysAllLogic {
	return &SysAllLogic{}
}

func (l *SysAllLogic) SysAll(ctx context.Context) (*v1.SysAllRes, error) {
	res := &v1.SysAllRes{}

	uptime, err := host.Uptime()
	if err == nil {
		res.Uptime = int64(uptime)
		res.TimeSinceUptime = time.Now().Add(-time.Duration(uptime) * time.Second).Format("2006-01-02 15:04:05")
	}

	procs, err := process.Pids()
	if err == nil {
		res.Procs = len(procs)
	}

	loadAvg, err := load.Avg()
	if err == nil {
		res.Load1 = loadAvg.Load1
		res.Load5 = loadAvg.Load5
		res.Load15 = loadAvg.Load15

		cpuCount, _ := cpu.Counts(true)
		if cpuCount > 0 {
			res.LoadUsagePercent = (loadAvg.Load1 / float64(cpuCount)) * 100
		}
	}

	cpuPercent, err := cpu.Percent(0, false)
	if err == nil && len(cpuPercent) > 0 {
		res.CpuPercent = cpuPercent
		res.CpuUsedPercent = cpuPercent[0]
	}

	cpuTimes, err := cpu.Times(false)
	if err == nil && len(cpuTimes) > 0 {
		total := cpuTimes[0].Total()
		idle := cpuTimes[0].Idle
		if total > 0 {
			res.CpuUsed = ((total - idle) / total) * 100
		}
		res.CpuDetailedPercent = []float64{
			cpuTimes[0].User,
			cpuTimes[0].Nice,
			cpuTimes[0].System,
			cpuTimes[0].Idle,
			cpuTimes[0].Iowait,
			cpuTimes[0].Irq,
			cpuTimes[0].Softirq,
			cpuTimes[0].Steal,
		}
	}

	memInfo, err := mem.VirtualMemory()
	if err == nil {
		res.MemoryTotal = memInfo.Total
		res.MemoryUsed = memInfo.Used
		res.MemoryFree = memInfo.Free
		res.MemoryShard = memInfo.Shared
		res.MemoryCache = memInfo.Cached
		res.MemoryAvailable = memInfo.Available
		res.MemoryUsedPercent = memInfo.UsedPercent
	}

	swapInfo, err := mem.SwapMemory()
	if err == nil {
		res.SwapMemoryTotal = swapInfo.Total
		res.SwapMemoryAvailable = swapInfo.Free
		res.SwapMemoryUsed = swapInfo.Used
		res.SwapMemoryUsedPercent = swapInfo.UsedPercent
	}

	ioCounters, err := disk.IOCounters()
	if err == nil {
		for _, io := range ioCounters {
			res.IoReadBytes += io.ReadBytes
			res.IoWriteBytes += io.WriteBytes
			res.IoCount += io.ReadCount + io.WriteCount
			res.IoReadTime += io.ReadTime
			res.IoWriteTime += io.WriteTime
		}
	}

	diskParts, err := disk.Partitions(false)
	if err == nil {
		res.DiskData = make([]v1.DiskInfo, 0)
		for _, part := range diskParts {
			diskUsage, err := disk.Usage(part.Mountpoint)
			if err == nil {
				diskInfo := v1.DiskInfo{
					Path:              part.Mountpoint,
					Type:              part.Fstype,
					Device:            part.Device,
					Total:             diskUsage.Total,
					Free:              diskUsage.Free,
					Used:              diskUsage.Used,
					UsedPercent:       diskUsage.UsedPercent,
					InodesTotal:       diskUsage.InodesTotal,
					InodesUsed:        diskUsage.InodesUsed,
					InodesFree:        diskUsage.InodesFree,
					InodesUsedPercent: diskUsage.InodesUsedPercent,
				}
				res.DiskData = append(res.DiskData, diskInfo)
			}
		}
	}

	netIO, err := net.IOCounters(false)
	if err == nil && len(netIO) > 0 {
		res.NetBytesSent = netIO[0].BytesSent
		res.NetBytesRecv = netIO[0].BytesRecv
	}

	res.ShotTime = time.Now().Format(time.RFC3339)

	return res, nil
}
