package system

import (
	"context"

	v1 "opskvm/api/system/v1"
)

func (c *ControllerV1) GetHostname(ctx context.Context, req *v1.GetHostnameReq) (res *v1.GetHostnameRes, err error) {
	hostname, err := c.sys.GetHostname(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetHostnameRes{
		Hostname: hostname,
	}, nil
}
