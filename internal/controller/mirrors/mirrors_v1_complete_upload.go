package mirrors

import (
	"context"
	"fmt"

	v1 "opskvm/api/mirrors/v1"
)

func (c *ControllerV1) CompleteUpload(ctx context.Context, req *v1.CompleteUploadReq) (res *v1.CompleteUploadRes, err error) {
	session, ok := c.mirrors.SVC.MirrorsChunkUploadService.GetSession(req.UploadId)
	if !ok {
		return nil, fmt.Errorf("上传会话不存在")
	}

	finalPath, err := c.mirrors.SVC.MirrorsChunkUploadService.CompleteUpload(ctx, req.UploadId)
	if err != nil {
		return nil, err
	}

	return &v1.CompleteUploadRes{
		UploadId:  req.UploadId,
		Id:        session.FileName,
		FileName:  session.FileName,
		FileSize:  session.FileSize,
		LocalPath: finalPath,
		URL:       "//" + session.FileName,
	}, nil
}
