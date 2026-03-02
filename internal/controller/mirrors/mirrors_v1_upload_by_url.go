package mirrors

import (
	"context"
	"encoding/base64"

	v1 "opskvm/api/mirrors/v1"
)

func (c *ControllerV1) UploadByUrl(ctx context.Context, req *v1.UploadByUrlReq) (res *v1.UploadByUrlRes, err error) {
	fileName := req.FileName
	if decoded, err := base64.StdEncoding.DecodeString(req.FileName); err == nil {
		fileName = string(decoded)
	}

	task, err := c.mirrors.SVC.MirrorsManagerService.UploadByUrl(req.Url, fileName)
	if err != nil {
		return nil, err
	}

	return &v1.UploadByUrlRes{
		UploadId:  task.UploadId,
		FileName:  task.FileName,
		FileSize:  task.FileSize,
		Status:    task.Status,
		LocalPath: task.LocalPath,
		URL:       task.URL,
	}, nil
}
