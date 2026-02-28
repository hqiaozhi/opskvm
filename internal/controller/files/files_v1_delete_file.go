package files

import (
	"context"

	v1 "opskvm/api/files/v1"
)

func (c *ControllerV1) DeleteFile(ctx context.Context, req *v1.DeleteFileReq) (res *v1.DeleteFileRes, err error) {
	deletedCount, failedIds, err := c.Files.SVC.FileManager.DeleteFiles(req.FileIds)
	if err != nil {
		return nil, err
	}

	return &v1.DeleteFileRes{
		DeletedCount: deletedCount,
		FailedIds:    failedIds,
	}, nil
}
