package files

import (
	"context"

	v1 "opskvm/api/files/v1"
)

func (c *ControllerV1) CancelUpload(ctx context.Context, req *v1.CancelUploadReq) (res *v1.CancelUploadRes, err error) {
	err = c.Files.SVC.ChunkUpload.CancelUpload(ctx, req.UploadId)
	if err != nil {
		return nil, err
	}

	return &v1.CancelUploadRes{}, nil
}
