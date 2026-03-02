package wake

import (
	"context"
	"opskvm/api/wake/v1"
)

func (c *ControllerV1) ListWol(ctx context.Context, req *v1.ListWolReq) (res *v1.ListWolRes, err error) {
	list, total, err := c.Wake.ListWol(ctx, req.Page, req.Limit, req.DeviceName, req.MacAddr)
	if err != nil {
		return nil, err
	}
	wolList := make([]v1.WolInfo, 0, len(list))
	for _, wol := range list {
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
		wolList = append(wolList, wolInfo)
	}
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 20
	}
	return &v1.ListWolRes{
		Total:   total,
		Page:    req.Page,
		Limit:   req.Limit,
		WolList: wolList,
	}, nil
}
