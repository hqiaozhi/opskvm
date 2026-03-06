package users

import (
	"context"

	v1 "opskvm/api/users/v1"
)

func (c *ControllerV1) Register(ctx context.Context, req *v1.RegisterReq) (res *v1.RegisterRes, err error) {
	userId, err := c.users.Register(ctx, req.Username, req.Password, req.Nickname, req.Email)
	if err != nil {
		return nil, err
	}
	return &v1.RegisterRes{Id: userId}, nil
}
