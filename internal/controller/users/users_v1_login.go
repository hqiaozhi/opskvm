package users

import (
	"context"
	"strconv"

	v1 "opskvm/api/users/v1"
)

func (c *ControllerV1) Login(ctx context.Context, req *v1.LoginReq) (res *v1.LoginRes, err error) {
	result, err := c.users.Login(ctx, req.Username, req.Password)
	if err != nil {
		return nil, err
	}

	token, err := c.users.JWT.GenerateAccessToken(strconv.Itoa(result.UserId), result.Username)
	if err != nil {
		return nil, err
	}

	return &v1.LoginRes{
		Token:        token,
		TotpRequired: result.TwoFactorEnabled,
		UserId:       result.UserId,
	}, nil
}
