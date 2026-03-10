package mirrors

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	v1 "opskvm/api/mirrors/v1"
)

func (c *ControllerV1) GetMsdStatus(ctx context.Context, req *v1.GetMsdStatusReq) (res *v1.GetMsdStatusRes, err error) {
	if c.mirrors.SVC.MSD == nil {
		return nil, gerror.New("MSD device not available")
	}

	path, err := c.mirrors.SVC.MSD.GetPath()
	if err != nil {
		return nil, err
	}
	return &v1.GetMsdStatusRes{
		Path: path,
	}, nil
}
