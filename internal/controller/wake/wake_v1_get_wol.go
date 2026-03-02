package wake

import (
	"context"
	v1 "opskvm/api/wake/v1"
)

func (c *ControllerV1) GetWol(ctx context.Context, req *v1.GetWolReq) (res *v1.GetWolRes, err error) {
	wol, err := c.Wake.GetWol(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	if wol == nil || wol.Id == 0 {
		return &v1.GetWolRes{
			Wol: v1.WolInfo{},
		}, nil
	}
	wolInfo := v1.WolInfo{
		Id:          wol.Id,
		DeviceName:  wol.DeviceName,
		MacAddr:     wol.MacAddr,
		BroadcastIp: wol.BroadcastIp,
		Port:        wol.Port,
		Remark:      wol.Remark,
	}
	if wol.CreatedAt != nil {
		wolInfo.CreatedAt = wol.CreatedAt.Format("Y-m-d H:i:s")
	}
	if wol.UpdatedAt != nil {
		wolInfo.UpdatedAt = wol.UpdatedAt.Format("Y-m-d H:i:s")
	}
	return &v1.GetWolRes{
		Wol: wolInfo,
	}, nil
}
