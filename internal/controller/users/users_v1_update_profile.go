package users

import (
	"context"

	v1 "opskvm/api/users/v1"

	"github.com/gogf/gf/v2/errors/gerror"
)

func (c *ControllerV1) UpdateProfile(ctx context.Context, req *v1.UpdateProfileReq) (res *v1.UpdateProfileRes, err error) {
	currentUserId := c.users.JWT.GetUserIdFromCtx(ctx)
	if currentUserId == 0 {
		return nil, gerror.New("未登录或token无效")
	}

	isAdmin, err := c.users.IsAdmin(ctx, currentUserId)
	if err != nil {
		return nil, err
	}

	var targetUserId int
	if isAdmin {
		if req.UserId == 0 {
			return nil, gerror.New("管理员操作需要指定用户ID")
		}
		targetUserId = req.UserId
	} else {
		targetUserId = currentUserId
	}

	err = c.users.UpdateProfile(ctx, targetUserId, req.Nickname, req.Email)
	if err != nil {
		return nil, err
	}

	return &v1.UpdateProfileRes{}, nil
}
