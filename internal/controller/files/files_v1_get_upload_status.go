package files

import (
	"context"

	v1 "opskvm/api/files/v1"
)

func (c *ControllerV1) GetUploadStatus(ctx context.Context, req *v1.GetUploadStatusReq) (res *v1.GetUploadStatusRes, err error) {
	session, err := c.Files.SVC.ChunkUpload.GetUploadStatus(req.UploadId)
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
		Path:          session.Path,
		FileName:      session.FileName,
		FileSize:      session.FileSize,
		ChunkSize:     session.ChunkSize,
		ChunkCount:    session.ChunkCount,
		UploadedCount: uploadedCount,
		UploadedSize:  uploadedSize,
	}, nil
}
