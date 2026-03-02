package system

import (
	"context"

	v1 "opskvm/api/system/v1"
)

func (c *ControllerV1) GetNetSpeed(ctx context.Context, req *v1.GetNetSpeedReq) (res *v1.GetNetSpeedRes, err error) {
	netSpeed, err := c.sys.GetNetSpeed(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetNetSpeedRes{
		Upload:   netSpeed.Upload,
		Download: netSpeed.Download,
	}, nil
}
