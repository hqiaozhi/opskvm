package wake

import (
	"context"
	"opskvm/api/wake/v1"
)

func (c *ControllerV1) DeleteWol(ctx context.Context, req *v1.DeleteWolReq) (res *v1.DeleteWolRes, err error) {
	err = c.Wake.DeleteWol(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.DeleteWolRes{}, nil
}
