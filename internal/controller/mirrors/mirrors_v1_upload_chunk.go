package mirrors

import (
	"context"
	"fmt"
	"io"

	"github.com/gogf/gf/v2/net/ghttp"

	v1 "opskvm/api/mirrors/v1"
)

func (c *ControllerV1) UploadChunk(ctx context.Context, req *v1.UploadChunkReq) (res *v1.UploadChunkRes, err error) {
	r := ghttp.RequestFromCtx(ctx)

	svc := c.mirrors.SVC.MirrorsChunkUploadService
	session, ok := svc.GetSession(req.UploadId)
	if !ok {
		return nil, fmt.Errorf("上传会话不存在: %s", req.UploadId)
	}

	if req.ChunkIndex < 0 || req.ChunkIndex >= session.ChunkCount {
		return nil, fmt.Errorf("分块索引无效: %d, 总分块数: %d", req.ChunkIndex, session.ChunkCount)
	}

	chunkData, err := io.ReadAll(r.Request.Body)
	if err != nil {
		return nil, fmt.Errorf("读取请求体失败: %v", err)
	}

	writtenSize, err := svc.UploadChunk(ctx, req.UploadId, req.ChunkIndex, chunkData)
	if err != nil {
		return nil, err
	}

	return &v1.UploadChunkRes{
		ChunkIndex: req.ChunkIndex,
		Offset:     writtenSize,
	}, nil
}
