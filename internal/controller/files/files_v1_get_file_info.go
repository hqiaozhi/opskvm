package files

import (
	"context"

	v1 "opskvm/api/files/v1"
)

func (c *ControllerV1) GetFileInfo(ctx context.Context, req *v1.GetFileInfoReq) (res *v1.GetFileInfoRes, err error) {
	info, err := c.Files.SVC.FileManager.GetFileInfo(req.FileId)
	if err != nil {
		return nil, err
	}

	return &v1.GetFileInfoRes{
		Name:      info.Name,
		Path:      info.Path,
		Size:      info.Size,
		IsDir:     info.IsDir,
		CreatedAt: info.CreatedAt,
	}, nil
}
