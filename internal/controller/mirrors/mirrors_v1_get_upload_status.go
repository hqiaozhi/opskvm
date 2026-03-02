package mirrors

import (
	"context"

	v1 "opskvm/api/mirrors/v1"
)

func (c *ControllerV1) GetUploadStatus(ctx context.Context, req *v1.GetUploadStatusReq) (res *v1.GetUploadStatusRes, err error) {
	session, err := c.mirrors.SVC.MirrorsChunkUploadService.GetUploadStatus(req.UploadId)
	if err != nil {
		return nil, err
	}

	uploadedCount := len(session.UploadedMap)
	var uploadedSize int64
	for _, size := range session.UploadedMap {
		uploadedSize += size
	}

	return &v1.GetUploadStatusRes{
		UploadID:      session.UploadID,
		FileName:      session.FileName,
		FileSize:      session.FileSize,
		ChunkSize:     session.ChunkSize,
		ChunkCount:    session.ChunkCount,
		UploadedCount: uploadedCount,
		UploadedSize:  uploadedSize,
	}, nil
}
