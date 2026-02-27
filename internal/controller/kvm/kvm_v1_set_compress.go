package kvm

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	v1 "opskvm/api/kvm/v1"
	"opskvm/internal/service/video"
)

func (c *ControllerV1) SetCompress(ctx context.Context, req *v1.SetCompressReq) (res *v1.SetCompressRes, err error) {

	// 获取当前流分发器实例
	streamer, ok := c.kvm.SVC.Streamer.(*video.MJPEGStreamer)
	if !ok {
		return res, gerror.New("Invalid streamer type")
	}

	// 根据请求启用或禁用压缩
	if req.Enabled {
		streamer.EnableCompression()
		// 如果请求中包含质量参数，则设置压缩质量
		if req.Quality > 0 {
			streamer.SetCompressQuality(req.Quality)
		}
	} else {
		streamer.DisableCompression()
	}

	// 获取当前压缩配置
	enabled, quality := streamer.GetCompressConfig()

	// 构建成功响应
	res = &v1.SetCompressRes{
		Enabled: enabled,
		Quality: quality,
	}

	return res, nil
}
