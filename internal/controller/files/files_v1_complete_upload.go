package files

import (
	"context"
	"fmt"

	v1 "opskvm/api/files/v1"
)

func (c *ControllerV1) CompleteUpload(ctx context.Context, req *v1.CompleteUploadReq) (res *v1.CompleteUploadRes, err error) {
	session, ok := c.Files.SVC.ChunkUpload.GetSession(req.UploadId)
	if !ok {
		return nil, fmt.Errorf("上传会话不存在")
	}

	finalPath, err := c.Files.SVC.ChunkUpload.CompleteUpload(ctx, req.UploadId)
	if err != nil {
		return nil, err
	}

	return &v1.CompleteUploadRes{
		UploadId: req.UploadId,
		FileId:   session.Path,
		FilePath: finalPath,
		FileName: session.FileName,
		FileSize: session.FileSize,
	}, nil
}
