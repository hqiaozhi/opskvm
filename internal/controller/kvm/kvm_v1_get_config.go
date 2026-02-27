package kvm

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	v1 "opskvm/api/kvm/v1"
)

func (c *ControllerV1) GetConfig(ctx context.Context, req *v1.GetConfigReq) (res *v1.GetConfigRes, err error) {
	cfg, err := c.kvm.SVC.Camera.GetConfig()
	if err != nil {
		return nil, gerror.Newf("Failed to get config: %v", err)
	}

	// 构建响应
	res = &v1.GetConfigRes{
		Width:  cfg.Width,
		Height: cfg.Height,
		FPS:    float64(cfg.FPS.N) / float64(cfg.FPS.D),
	}
	return res, nil
}
