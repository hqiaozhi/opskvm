package files

import (
	"context"

	v1 "opskvm/api/files/v1"
)

func (c *ControllerV1) GetStorageInfo(ctx context.Context, req *v1.GetStorageInfoReq) (res *v1.GetStorageInfoRes, err error) {
	info, err := c.Files.SVC.FileManager.GetStorageInfo()
	if err != nil {
		return nil, err
	}

	return &v1.GetStorageInfoRes{
		TotalSpace:      info.TotalSpace,
		UsedSpace:       info.UsedSpace,
		AvailableSpace:  info.AvailableSpace,
		UsedPercent:     info.UsedPercent,
	}, nil
}
