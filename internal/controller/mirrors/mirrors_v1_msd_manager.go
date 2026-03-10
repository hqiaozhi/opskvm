package mirrors

import (
	"context"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"

	v1 "opskvm/api/mirrors/v1"
)

func (c *ControllerV1) MsdManager(ctx context.Context, req *v1.MsdManagerReq) (res *v1.MsdManagerRes, err error) {
	if c.mirrors.SVC.MSD == nil {
		return nil, gerror.New("MSD device not available")
	}

	// 挂载模式: 0(Flash)/1(CD/DVD)/2(取消挂载)
	switch req.Mode {
	case "0":
		return nil, c.mirrors.SVC.MSD.Bind(req.Path, "0")
	case "1":
		return nil, c.mirrors.SVC.MSD.Bind(req.Path, "1")
	case "2":
		return nil, c.mirrors.SVC.MSD.Remove()
	default:
		return nil, gerror.NewCode(gcode.CodeInvalidParameter)
	}
}
