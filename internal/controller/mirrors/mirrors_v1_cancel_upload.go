package mirrors

import (
	"context"

	v1 "opskvm/api/mirrors/v1"
)

func (c *ControllerV1) CancelUpload(ctx context.Context, req *v1.CancelUploadReq) (res *v1.CancelUploadRes, err error) {
	err = c.mirrors.SVC.MirrorsChunkUploadService.CancelUpload(ctx, req.UploadId)
	if err != nil {
		return nil, err
	}

	return &v1.CancelUploadRes{}, nil
}
