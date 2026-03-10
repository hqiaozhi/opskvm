package kvm

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	v1 "opskvm/api/kvm/v1"
)

func (c *ControllerV1) TurnOn(ctx context.Context, req *v1.TurnOnReq) (res *v1.TurnOnRes, err error) {
	// 检查摄像头是否可用
	if c.kvm.SVC.Camera == nil {
		return nil, gerror.New("Camera device not available")
	}

	// 打开摄像头
	err = c.kvm.SVC.Camera.TurnOn()
	if err != nil {
		return nil, gerror.Newf("Failed to turn on camera: %v", err)
	}

	return res, nil
}
