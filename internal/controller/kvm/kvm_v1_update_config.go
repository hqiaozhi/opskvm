package kvm

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	v1 "opskvm/api/kvm/v1"
)

func (c *ControllerV1) UpdateConfig(ctx context.Context, req *v1.UpdateConfigReq) (res *v1.UpdateConfigRes, err error) {
	// 检查设备是否可用
	if c.kvm.SVC.Camera == nil {
		return nil, gerror.New("Camera device not available")
	}
	if c.kvm.SVC.Streamer == nil {
		return nil, gerror.New("Streamer not available")
	}

	// 暂停流分发，避免配置更新冲突（延长暂停时间，适配设备重启）
	c.kvm.SVC.Streamer.Pause()
	defer func() {
		// 恢复流分发前短暂延迟，确保设备完全启动
		time.Sleep(200 * time.Millisecond)
		c.kvm.SVC.Streamer.Resume()
	}()

	// 更新摄像头配置
	actualCfg, err := c.kvm.SVC.Camera.UpdateConfig(int(req.Width), int(req.Height), uint32(req.FPS))
	if err != nil {
		return nil, gerror.Newf("Failed to update config: %v", err)
	}

	// 构建成功响应
	res = &v1.UpdateConfigRes{
		Width:  actualCfg.Width,
		Height: actualCfg.Height,
		FPS:    actualCfg.FPS.N / actualCfg.FPS.D,
	}

	return res, nil
}
