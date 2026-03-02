package v1

import "github.com/gogf/gf/v2/frame/g"

type RestartSysReq struct {
	g.Meta `path:"system/restart" method:"post" sm:"重启系统" tags:"系统管理"`
}
type RestartSysRes struct{}

type GetHostnameReq struct {
	g.Meta `path:"system/gethostname" method:"get" sm:"获取主机名" tags:"系统管理"`
}
type GetHostnameRes struct {
	Hostname string `json:"hostname" dc:"主机名"`
}

type GetNetSpeedReq struct {
	g.Meta `path:"system/getnetspeed" method:"get" sm:"获取网络速度" tags:"系统管理"`
}
type GetNetSpeedRes struct {
	Upload   int64 `json:"upload" dc:"上传速度 (bytes/s)"`
	Download int64 `json:"download" dc:"下载速度 (bytes/s)"`
}
