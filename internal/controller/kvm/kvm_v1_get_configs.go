package kvm

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/korandiz/v4l/fmt/mjpeg"

	v1 "opskvm/api/kvm/v1"
)

func (c *ControllerV1) GetConfigs(ctx context.Context, req *v1.GetConfigsReq) (res *v1.GetConfigsRes, err error) {
	// 检查摄像头是否可用
	if c.kvm.SVC.Camera == nil {
		return nil, gerror.New("Camera device not available")
	}

	supportedCfgs, err := c.kvm.SVC.Camera.ListConfigs()
	if err != nil {
		return nil, gerror.Newf("Failed to get supported configs: %v", err)
	}

	// 过滤并转换为简化的Config格式（只返回MJPEG格式的配置）
	var configs []v1.Config
	for _, cfg := range supportedCfgs {
		if cfg.Format == mjpeg.FourCC {
			configs = append(configs, v1.Config{
				Width:  cfg.Width,
				Height: cfg.Height,
				FPS:    float64(cfg.FPS.N) / float64(cfg.FPS.D),
			})
		}
	}

	// 构建响应
	res = &v1.GetConfigsRes{
		Configs: configs,
	}
	return res, nil
}
