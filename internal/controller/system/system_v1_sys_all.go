package system

import (
	"context"

	v1 "opskvm/api/system/v1"
	"opskvm/internal/logic/system"
)

func (c *ControllerV1) SysAll(ctx context.Context, req *v1.SysAllReq) (res *v1.SysAllRes, err error) {
	return system.NewSysAllLogic().SysAll(ctx)
}
