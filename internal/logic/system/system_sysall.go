package system

import (
	"context"
	"os"
	"os/exec"
	"strconv"
	"strings"
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
		res.CpuTotal = cpuCount
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
		t := cpuTimes[0]
		total := t.User + t.Nice + t.System + t.Idle + t.Iowait + t.Irq + t.Softirq + t.Steal
		idle := t.Idle
		if total > 0 {
			res.CpuUsed = ((total - idle) / total) * 100
		}
		res.CpuDetailedPercent = []float64{
			t.User,
			t.Nice,
			t.System,
			t.Idle,
			t.Iowait,
			t.Irq,
			t.Softirq,
			t.Steal,
		}
	}

	cpuTemp, err := getCPUTemperature()
	if err == nil {
		res.CpuTemperature = cpuTemp
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
			if strings.HasPrefix(part.Device, "/dev/loop") {
				continue
			}
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

func getCPUTemperature() (float64, error) {
	temp := getCPUTemperatureFromSysfs()
	if temp > 0 {
		return temp, nil
	}

	temp = getCPUTemperatureFromHwmon()
	if temp > 0 {
		return temp, nil
	}

	temp = getCPUTemperatureFromSensors()
	if temp > 0 {
		return temp, nil
	}

	temp = getCPUTemperatureFromThermalZone()
	if temp > 0 {
		return temp, nil
	}

	return 0, nil
}

func getCPUTemperatureFromSysfs() float64 {
	paths := []string{
		"/sys/class/thermal/thermal_zone0/temp",
		"/sys/devices/virtual/thermal/thermal_zone0/temp",
	}

	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err == nil {
			temp, err := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
			if err == nil {
				return temp / 1000.0
			}
		}
	}
	return 0
}

func getCPUTemperatureFromHwmon() float64 {
	hwmonDir := "/sys/class/hwmon"
	entries, err := os.ReadDir(hwmonDir)
	if err != nil {
		return 0
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		namePath := hwmonDir + "/" + entry.Name() + "/name"
		nameData, err := os.ReadFile(namePath)
		if err != nil {
			continue
		}

		name := strings.TrimSpace(string(nameData))
		if name == "cpu" || name == "coretemp" || strings.Contains(name, "cpu_") || strings.Contains(name, "k10temp") || strings.Contains(name, "acpitz") {
			for i := 0; i < 10; i++ {
				tempPath := hwmonDir + "/" + entry.Name() + "/temp" + strconv.Itoa(i) + "_input"
				data, err := os.ReadFile(tempPath)
				if err == nil {
					temp, err := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
					if err == nil && temp > 0 && temp < 150000 {
						return temp / 1000.0
					}
				}
			}
		}
	}
	return 0
}

func getCPUTemperatureFromSensors() float64 {
	cmd := exec.Command("sensors", "-j")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "temp1_input") || strings.Contains(line, "Package") {
			parts := strings.Fields(line)
			for _, part := range parts {
				temp, err := strconv.ParseFloat(strings.TrimSuffix(part, ","), 64)
				if err == nil && temp > 0 && temp < 150 {
					return temp
				}
			}
		}
	}

	for _, line := range lines {
		if strings.Contains(line, "Core 0") || strings.Contains(line, "Core 1") || strings.Contains(line, "CPU") {
			parts := strings.Fields(line)
			for _, part := range parts {
				temp, err := strconv.ParseFloat(strings.TrimSuffix(part, "°C"), 64)
				if err == nil && temp > 0 && temp < 150 {
					return temp
				}
			}
		}
	}

	return 0
}

func getCPUTemperatureFromThermalZone() float64 {
	thermalDir := "/sys/class/thermal"
	entries, err := os.ReadDir(thermalDir)
	if err != nil {
		return 0
	}

	var maxTemp float64
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		typePath := thermalDir + "/" + entry.Name() + "/type"
		typeData, err := os.ReadFile(typePath)
		if err != nil {
			continue
		}

		typeStr := strings.TrimSpace(string(typeData))
		if strings.Contains(typeStr, "cpu") || strings.Contains(typeStr, "x86_pkg_temp") || strings.Contains(typeStr, "acpitz") {
			tempPath := thermalDir + "/" + entry.Name() + "/temp"
			data, err := os.ReadFile(tempPath)
			if err == nil {
				temp, err := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
				if err == nil && temp > 0 {
					tempC := temp / 1000.0
					if tempC > maxTemp {
						maxTemp = tempC
					}
				}
			}
		}
	}

	return maxTemp
}
