package mirrors

import (
	"context"

	v1 "opskvm/api/mirrors/v1"
)

func (c *ControllerV1) GetUrlUploadStatus(ctx context.Context, req *v1.GetUrlUploadStatusReq) (res *v1.GetUrlUploadStatusRes, err error) {
	task, err := c.mirrors.SVC.MirrorsManagerService.GetUrlUploadStatus(req.UploadId)
	if err != nil {
		return nil, err
	}

	return &v1.GetUrlUploadStatusRes{
		UploadId:  task.UploadId,
		FileName:  task.FileName,
		FileSize:  task.FileSize,
		Status:    task.Status,
		LocalPath: task.LocalPath,
		URL:       task.URL,
		Error:     task.Error,
	}, nil
}
