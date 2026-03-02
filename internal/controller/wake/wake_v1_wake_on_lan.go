package wake

import (
	"context"
	v1 "opskvm/api/wake/v1"
)

func (c *ControllerV1) WakeOnLan(ctx context.Context, req *v1.WakeOnLanReq) (res *v1.WakeOnLanRes, err error) {
	err = c.Wake.WakeOnLan(ctx, req.BroadcastIp, req.Port, req.MacAddr)
	if err != nil {
		return &v1.WakeOnLanRes{
			Success: false,
			Message: err.Error(),
		}, nil
	}
	return &v1.WakeOnLanRes{
		Success: true,
		Message: "Magic packet sent successfully",
	}, nil
}
