package files

import (
	"context"

	v1 "opskvm/api/files/v1"
)

func (c *ControllerV1) ListFiles(ctx context.Context, req *v1.ListFilesReq) (res *v1.ListFilesRes, err error) {
	total, fileList, err := c.Files.SVC.FileManager.ListFiles(req.Path, req.Page, req.Limit)
	if err != nil {
		return nil, err
	}

	var list []v1.FileInfo
	for _, f := range fileList {
		list = append(list, v1.FileInfo{
			Id:        f.Id,
			Name:      f.Name,
			Path:      f.Path,
			RelPath:   f.RelPath,
			Size:      f.Size,
			IsDir:     f.IsDir,
			CreatedAt: f.CreatedAt,
			UpdatedAt: f.UpdatedAt,
		})
	}

	return &v1.ListFilesRes{
		Total:    total,
		Page:     req.Page,
		Limit:    req.Limit,
		FileList: list,
	}, nil
}
