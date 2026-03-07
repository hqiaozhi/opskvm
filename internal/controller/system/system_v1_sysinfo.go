package system

import (
	"context"

	v1 "opskvm/api/system/v1"
)

func (c *ControllerV1) SysInfo(ctx context.Context, req *v1.SysInfoReq) (res *v1.SysInfoRes, err error) {
	info, err := c.sys.SysInfo(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.SysInfoRes{
		Hostname: info.Hostname,
		Distro:   info.Distro,
		Kernel:   info.Kernel,
		Arch:     info.Arch,
		HostIP:   info.HostIP,
		BootTime: info.BootTime,
		Uptime:   info.Uptime,
	}, nil
}
