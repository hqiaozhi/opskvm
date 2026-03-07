// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package system

import (
	"context"

	"opskvm/api/system/v1"
)

type ISystemV1 interface {
	SysAll(ctx context.Context, req *v1.SysAllReq) (res *v1.SysAllRes, err error)
	RestartSys(ctx context.Context, req *v1.RestartSysReq) (res *v1.RestartSysRes, err error)
	SysInfo(ctx context.Context, req *v1.SysInfoReq) (res *v1.SysInfoRes, err error)
	GetNetSpeed(ctx context.Context, req *v1.GetNetSpeedReq) (res *v1.GetNetSpeedRes, err error)
}
