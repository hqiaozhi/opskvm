package wake

import (
	"context"
	"opskvm/api/wake/v1"
)

func (c *ControllerV1) CreateWol(ctx context.Context, req *v1.CreateWolReq) (res *v1.CreateWolRes, err error) {
	id, err := c.Wake.CreateWol(ctx, req.DeviceName, req.MacAddr, req.BroadcastIp, req.Port, req.Remark)
	if err != nil {
		return nil, err
	}
	return &v1.CreateWolRes{
		Id: int(id),
	}, nil
}
