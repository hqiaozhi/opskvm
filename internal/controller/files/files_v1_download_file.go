package files

import (
	"context"

	"github.com/gogf/gf/v2/net/ghttp"

	v1 "opskvm/api/files/v1"
)

func (c *ControllerV1) DownloadFile(ctx context.Context, req *v1.DownloadFileReq) (res *v1.DownloadFileRes, err error) {
	r := ghttp.RequestFromCtx(ctx)
	err = c.Files.SVC.FileManager.DownloadFile(req.FileId, r.Response.Writer)
	if err != nil {
		r.Response.WriteStatus(404, "文件不存在")
		return nil, err
	}

	return nil, nil
}
