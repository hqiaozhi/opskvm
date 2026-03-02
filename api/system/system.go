// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package system

import (
	"context"

	"opskvm/api/system/v1"
)

type ISystemV1 interface {
	RestartSys(ctx context.Context, req *v1.RestartSysReq) (res *v1.RestartSysRes, err error)
	GetHostname(ctx context.Context, req *v1.GetHostnameReq) (res *v1.GetHostnameRes, err error)
	GetNetSpeed(ctx context.Context, req *v1.GetNetSpeedReq) (res *v1.GetNetSpeedRes, err error)
}
