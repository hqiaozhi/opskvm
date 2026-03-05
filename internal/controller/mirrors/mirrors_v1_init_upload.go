package mirrors

import (
	"context"
	"encoding/base64"

	v1 "opskvm/api/mirrors/v1"
	mirrorssvc "opskvm/internal/service/mirrors"
)

func (c *ControllerV1) InitUpload(ctx context.Context, req *v1.InitUploadReq) (res *v1.InitUploadRes, err error) {
	var Files []mirrorssvc.FileMeta
	for _, f := range req.Files {
		fileName := f.FileName
		if decoded, err := base64.StdEncoding.DecodeString(f.FileName); err == nil {
			fileName = string(decoded)
		}
		Files = append(Files, mirrorssvc.FileMeta{
			FileName: fileName,
			Size:     f.Size,
			Md5:      f.Md5,
		})
	}

	sessions, err := c.mirrors.SVC.MirrorsChunkUploadService.InitUpload(Files, req.ChunkSize)
	if err != nil {
		return nil, err
	}

	var uploads []v1.UploadSessionInfo
	for _, s := range sessions {
		uploads = append(uploads, v1.UploadSessionInfo{
			UploadID:   s.UploadID,
			FileName:   s.FileName,
			FileSize:   s.FileSize,
			ChunkSize:  s.ChunkSize,
			ChunkCount: s.ChunkCount,
			FileMd5:    s.FileMd5,
			SkipUpload: s.SkipUpload,
		})
	}

	return &v1.InitUploadRes{
		Uploads: uploads,
	}, nil
}
