package v1

import "github.com/gogf/gf/v2/frame/g"

type SysAllReq struct {
	g.Meta `path:"system/sysall" method:"get" sm:"获取系统状态" tags:"系统管理"`
}
type SysAllRes struct {
	Uptime                int64       `json:"uptime" dc:"运行时间(秒)"`
	TimeSinceUptime       string      `json:"timeSinceUptime" dc:"系统启动时间"`
	Procs                 int         `json:"procs" dc:"进程数"`
	Load1                 float64     `json:"load1" dc:"1分钟负载"`
	Load5                 float64     `json:"load5" dc:"5分钟负载"`
	Load15                float64     `json:"load15" dc:"15分钟负载"`
	LoadUsagePercent      float64     `json:"loadUsagePercent" dc:"负载使用百分比"`
	CpuPercent            []float64   `json:"cpuPercent" dc:"CPU使用百分比数组"`
	CpuUsedPercent        float64     `json:"cpuUsedPercent" dc:"CPU使用百分比"`
	CpuUsed               float64     `json:"cpuUsed" dc:"CPU使用率"`
	CpuTotal              int         `json:"cpuTotal" dc:"CPU核心数"`
	CpuDetailedPercent    []float64   `json:"cpuDetailedPercent" dc:"详细CPU百分比"`
	MemoryTotal           uint64      `json:"memoryTotal" dc:"内存总量"`
	MemoryUsed            uint64      `json:"memoryUsed" dc:"已用内存"`
	MemoryFree            uint64      `json:"memoryFree" dc:"空闲内存"`
	MemoryShard           uint64      `json:"memoryShard" dc:"共享内存"`
	MemoryCache           uint64      `json:"memoryCache" dc:"缓存内存"`
	MemoryAvailable       uint64      `json:"memoryAvailable" dc:"可用内存"`
	MemoryUsedPercent     float64     `json:"memoryUsedPercent" dc:"内存使用百分比"`
	SwapMemoryTotal       uint64      `json:"swapMemoryTotal" dc:"交换内存总量"`
	SwapMemoryAvailable   uint64      `json:"swapMemoryAvailable" dc:"可用交换内存"`
	SwapMemoryUsed        uint64      `json:"swapMemoryUsed" dc:"已用交换内存"`
	SwapMemoryUsedPercent float64     `json:"swapMemoryUsedPercent" dc:"交换内存使用百分比"`
	IoReadBytes           uint64      `json:"ioReadBytes" dc:"读取字节数"`
	IoWriteBytes          uint64      `json:"ioWriteBytes" dc:"写入字节数"`
	IoCount               uint64      `json:"ioCount" dc:"IO操作次数"`
	IoReadTime            uint64      `json:"ioReadTime" dc:"读取时间"`
	IoWriteTime           uint64      `json:"ioWriteTime" dc:"写入时间"`
	DiskData              []DiskInfo  `json:"diskData" dc:"磁盘信息"`
	NetBytesSent          uint64      `json:"netBytesSent" dc:"发送字节数"`
	NetBytesRecv          uint64      `json:"netBytesRecv" dc:"接收字节数"`
	GpuData               interface{} `json:"gpuData" dc:"GPU数据"`
	XpuData               interface{} `json:"xpuData" dc:"XPU数据"`
	TopCPUItems           interface{} `json:"topCPUItems" dc:"CPU占用排行"`
	TopMemItems           interface{} `json:"topMemItems" dc:"内存占用排行"`
	ShotTime              string      `json:"shotTime" dc:"获取时间"`
}

type DiskInfo struct {
	Path              string  `json:"path" dc:"挂载点"`
	Type              string  `json:"type" dc:"文件系统类型"`
	Device            string  `json:"device" dc:"设备名"`
	Total             uint64  `json:"total" dc:"总容量"`
	Free              uint64  `json:"free" dc:"可用空间"`
	Used              uint64  `json:"used" dc:"已用空间"`
	UsedPercent       float64 `json:"usedPercent" dc:"使用百分比"`
	InodesTotal       uint64  `json:"inodesTotal" dc:"inode总数"`
	InodesUsed        uint64  `json:"inodesUsed" dc:"已用inode"`
	InodesFree        uint64  `json:"inodesFree" dc:"可用inode"`
	InodesUsedPercent float64 `json:"inodesUsedPercent" dc:"inode使用百分比"`
}
