package users

import (
	"context"

	v1 "opskvm/api/users/v1"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

func (c *ControllerV1) Register(ctx context.Context, req *v1.RegisterReq) (res *v1.RegisterRes, err error) {
	enroll := g.Cfg().MustGetWithCmd(ctx, `enroll`)
	if !enroll.Bool() {
		return nil, gerror.New("Registration prohibited")
	}

	userId, err := c.users.Register(ctx, req.Username, req.Password, req.Nickname, req.Email)
	if err != nil {
		return nil, err
	}
	return &v1.RegisterRes{Id: userId}, nil
}
