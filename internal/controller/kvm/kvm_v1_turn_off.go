package kvm

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	v1 "opskvm/api/kvm/v1"
)

func (c *ControllerV1) TurnOff(ctx context.Context, req *v1.TurnOffReq) (res *v1.TurnOffRes, err error) {
	// 关闭摄像头
	err = c.kvm.SVC.Camera.TurnOff()
	if err != nil {
		return nil, gerror.Newf("Failed to turn off camera: %v", err)
	}

	// 构建成功响应
	return res, nil
}
