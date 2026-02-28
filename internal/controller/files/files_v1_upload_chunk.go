package files

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/gogf/gf/v2/net/ghttp"

	v1 "opskvm/api/files/v1"
)

func (c *ControllerV1) UploadChunk(ctx context.Context, req *v1.UploadChunkReq) (res *v1.UploadChunkRes, err error) {
	r := ghttp.RequestFromCtx(ctx)

	svc := c.Files.SVC.ChunkUpload
	session, ok := svc.GetSession(req.UploadId)
	if !ok {
		return nil, fmt.Errorf("上传会话不存在: %s", req.UploadId)
	}

	if req.ChunkIndex < 0 || req.ChunkIndex >= session.ChunkCount {
		return nil, fmt.Errorf("分块索引无效: %d, 总分块数: %d", req.ChunkIndex, session.ChunkCount)
	}

	uploadDir := svc.GetUploadDir(req.UploadId)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("创建上传目录失败: %v", err)
	}

	chunkPath := filepath.Join(uploadDir, fmt.Sprintf("chunk_%05d", req.ChunkIndex))

	file, err := os.Create(chunkPath)
	if err != nil {
		return nil, fmt.Errorf("创建分块文件失败: %v", err)
	}
	defer file.Close()

	writtenSize, err := io.Copy(file, r.Request.Body)
	if err != nil {
		os.Remove(chunkPath)
		return nil, fmt.Errorf("写入分块数据失败: %v", err)
	}

	svc.RecordChunk(req.UploadId, req.ChunkIndex, writtenSize)

	fmt.Printf("[UploadChunk] uploadId=%s, chunkIndex=%d, size=%d\n", req.UploadId, req.ChunkIndex, writtenSize)

	return &v1.UploadChunkRes{
		ChunkIndex: req.ChunkIndex,
		Offset:     writtenSize,
	}, nil
}
