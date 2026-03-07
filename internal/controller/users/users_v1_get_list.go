package users

import (
	"context"

	v1 "opskvm/api/users/v1"

	"github.com/gogf/gf/v2/errors/gerror"
)

func (c *ControllerV1) GetUserList(ctx context.Context, req *v1.GetUserListReq) (res *v1.GetUserListRes, err error) {
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

	total, userList, err := c.users.GetUserList(ctx, req.Page, req.Limit)
	if err != nil {
		return nil, err
	}

	var list []v1.GetUserInfoRes
	for _, user := range userList {
		list = append(list, v1.GetUserInfoRes{
			Id:               user.Id,
			Username:         user.Username,
			Nickname:         user.Nickname,
			Email:            user.Email,
			IsAdmin:          user.IsAdmin,
			TwoFactorEnabled: user.TwoFactorEnabled,
			CreatedAt:        user.CreatedAt,
		})
	}

	return &v1.GetUserListRes{
		Total: total,
		List:  list,
	}, nil
}
