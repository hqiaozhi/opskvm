package video

import (
	"opskvm/internal/svc"
)

// VideoHandler 视频处理器
type VideoHandler struct {
	svcCtx *svc.SvcContext
}

// NewVideoHandler 创建视频处理器实例
func NewVideoHandler(svcCtx *svc.SvcContext) *VideoHandler {
	return &VideoHandler{
		svcCtx: svcCtx,
	}
}

