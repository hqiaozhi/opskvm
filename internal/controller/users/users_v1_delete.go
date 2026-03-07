package users

import (
	"context"

	v1 "opskvm/api/users/v1"

	"github.com/gogf/gf/v2/errors/gerror"
)

func (c *ControllerV1) DeleteUser(ctx context.Context, req *v1.DeleteUserReq) (res *v1.DeleteUserRes, err error) {
	currentUserId := c.users.JWT.GetUserIdFromCtx(ctx)
	if currentUserId == 0 {
		return nil, gerror.New("unauthorized")
	}

	isAdmin, err := c.users.IsAdmin(ctx, currentUserId)
	if err != nil {
		return nil, err
	}

	if !isAdmin {
		return nil, gerror.New("permission denied, admin only")
	}

	targetUser, err := c.users.GetUserById(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	if targetUser.IsAdmin == 1 {
		return nil, gerror.New("cannot delete admin user")
	}

	err = c.users.DeleteUser(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	return &v1.DeleteUserRes{}, nil
}
