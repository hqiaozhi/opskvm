package users

import (
	"context"
	"strconv"

	v1 "opskvm/api/users/v1"
)

func (c *ControllerV1) Login(ctx context.Context, req *v1.LoginReq) (res *v1.LoginRes, err error) {
	user_id, err := c.users.Login(ctx, req.Username, req.Password)
	if err != nil {
		return nil, err
	}
	token, err := c.users.JWT.GenerateAccessToken(strconv.Itoa(user_id), req.Username)
	if err != nil {
		return nil, err
	}
	return &v1.LoginRes{Token: token}, nil
}
