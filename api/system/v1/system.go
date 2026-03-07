package v1

import "github.com/gogf/gf/v2/frame/g"

type RestartSysReq struct {
	g.Meta `path:"system/restart" method:"post" sm:"重启系统" tags:"系统管理"`
}
type RestartSysRes struct{}

type SysInfoReq struct {
	g.Meta `path:"system/sysinfo" method:"get" sm:"获取系统信息" tags:"系统管理"`
}
type SysInfoRes struct {
	Hostname string `json:"hostname" dc:"主机名称"`
	Distro   string `json:"distro" dc:"发行版本"`
	Kernel   string `json:"kernel" dc:"内核版本"`
	Arch     string `json:"arch" dc:"系统类型"`
	HostIP   string `json:"hostIP" dc:"主机地址"`
	BootTime string `json:"bootTime" dc:"启动时间"`
	Uptime   string `json:"uptime" dc:"运行时间"`
	CpuModel string `json:"cpuModel" dc:"CPU型号"`
}

type GetNetSpeedReq struct {
	g.Meta `path:"system/getnetspeed" method:"get" sm:"获取网络速度" tags:"系统管理"`
}
type GetNetSpeedRes struct {
	Upload   int64 `json:"upload" dc:"上传速度 (bytes/s)"`
	Download int64 `json:"download" dc:"下载速度 (bytes/s)"`
}
