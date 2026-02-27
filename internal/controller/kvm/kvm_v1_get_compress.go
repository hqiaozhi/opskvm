package kvm

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	v1 "opskvm/api/kvm/v1"
	"opskvm/internal/service/video"
)

func (c *ControllerV1) GetCompress(ctx context.Context, req *v1.GetCompressReq) (res *v1.GetCompressRes, err error) {
	// 获取当前流分发器实例
	streamer, ok := c.kvm.SVC.Streamer.(*video.MJPEGStreamer)
	if !ok {
		return res, gerror.New("Invalid streamer type")
	}

	// 获取当前压缩配置
	enabled, quality := streamer.GetCompressConfig()

	// 构建成功响应
	res = &v1.GetCompressRes{
		Enabled: enabled,
		Quality: quality,
	}

	return res, nil
}
