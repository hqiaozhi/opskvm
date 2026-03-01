package system

import (
	"context"

	v1 "opskvm/api/system/v1"
)

func (c *ControllerV1) RestartSys(ctx context.Context, req *v1.RestartSysReq) (res *v1.RestartSysRes, err error) {
	err = c.sys.RestartSys(ctx)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
