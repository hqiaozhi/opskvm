package wake

import (
	"context"
	"opskvm/api/wake/v1"
)

func (c *ControllerV1) UpdateWol(ctx context.Context, req *v1.UpdateWolReq) (res *v1.UpdateWolRes, err error) {
	err = c.Wake.UpdateWol(ctx, req.Id, req.DeviceName, req.MacAddr, req.BroadcastIp, req.Port, req.Remark)
	if err != nil {
		return nil, err
	}
	return &v1.UpdateWolRes{}, nil
}
