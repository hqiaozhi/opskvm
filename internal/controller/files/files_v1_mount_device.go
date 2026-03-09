package files

import (
	"context"

	v1 "opskvm/api/files/v1"
)

func (c *ControllerV1) MountDevice(ctx context.Context, req *v1.MountDeviceReq) (res *v1.MountDeviceRes, err error) {
	mountPoint, err := c.Files.MountDevice(ctx, req.DevPath)
	if err != nil {
		return nil, err
	}

	return &v1.MountDeviceRes{
		DevPath:    req.DevPath,
		MountPoint: mountPoint,
	}, nil
}
