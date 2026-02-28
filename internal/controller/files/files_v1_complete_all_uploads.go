package files

import (
	"context"

	v1 "opskvm/api/files/v1"
)

func (c *ControllerV1) CompleteAllUploads(ctx context.Context, req *v1.CompleteAllUploadsReq) (res *v1.CompleteAllUploadsRes, err error) {
	completedCount, completedSessions, err := c.Files.SVC.ChunkUpload.CompleteAllUploads(ctx)
	if err != nil {
		return nil, err
	}

	var results []v1.CompleteUploadRes
	for _, session := range completedSessions {
		results = append(results, v1.CompleteUploadRes{
			UploadId: session.UploadID,
			FileId:   session.Path,
			FileName: session.FileName,
			FileSize: session.FileSize,
		})
	}

	return &v1.CompleteAllUploadsRes{
		CompletedCount: completedCount,
		Results:        results,
	}, nil
}
