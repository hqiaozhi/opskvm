package users

import (
	"context"

	v1 "opskvm/api/users/v1"
)

func (c *ControllerV1) Logout(ctx context.Context, req *v1.LogoutReq) (res *v1.LogoutRes, err error) {
	return nil, nil
}
