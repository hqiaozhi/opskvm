package files

import (
	"context"
	"encoding/base64"

	v1 "opskvm/api/files/v1"
	"opskvm/internal/service/files"
)

func (c *ControllerV1) InitUpload(ctx context.Context, req *v1.InitUploadReq) (res *v1.InitUploadRes, err error) {
	var fileMetas []files.FileMeta
	for _, f := range req.Files {
		path := f.Path

		if decoded, err := base64.StdEncoding.DecodeString(f.Path); err == nil {
			path = string(decoded)
		}

		fileMetas = append(fileMetas, files.FileMeta{
			Path: path,
			Size: f.Size,
			Md5:  f.Md5,
		})
	}

	sessions, err := c.Files.SVC.ChunkUpload.InitUpload(fileMetas, req.ChunkSize)
	if err != nil {
		return nil, err
	}

	var uploads []v1.UploadSessionInfo
	for _, s := range sessions {
		uploads = append(uploads, v1.UploadSessionInfo{
			UploadID:   s.UploadID,
			Path:       s.Path,
			FileName:   s.FileName,
			FileSize:   s.FileSize,
			ChunkSize:  s.ChunkSize,
			ChunkCount: s.ChunkCount,
			IsDir:      s.IsDir,
			FileMd5:    s.FileMd5,
			SkipUpload: s.SkipUpload,
		})
	}

	return &v1.InitUploadRes{
		Uploads: uploads,
	}, nil
}
