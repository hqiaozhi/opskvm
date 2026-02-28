package files

import (
	"context"

	v1 "opskvm/api/files/v1"
)

func (c *ControllerV1) CreateDirectory(ctx context.Context, req *v1.CreateDirectoryReq) (res *v1.CreateDirectoryRes, err error) {
	err = c.Files.SVC.FileManager.CreateDirectory(req.Path)
	if err != nil {
		return nil, err
	}

	return &v1.CreateDirectoryRes{
		Path: req.Path,
	}, nil
}
