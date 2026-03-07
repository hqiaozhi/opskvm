package users

import (
	"context"

	v1 "opskvm/api/users/v1"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

func (c *ControllerV1) GetUserInfo(ctx context.Context, req *v1.GetUserInfoReq) (res *v1.GetUserInfoRes, err error) {
	r := ghttp.RequestFromCtx(ctx)
	currentUserId := r.GetParam("userid").Int()

	isAdmin, err := c.users.IsAdmin(ctx, currentUserId)
	if err != nil {
		return nil, err
	}

	if !isAdmin && currentUserId != req.UserId {
		return nil, gerror.New("permission denied")
	}

	userInfo, err := c.users.GetUserInfo(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	return &v1.GetUserInfoRes{
		Id:               userInfo.Id,
		Username:         userInfo.Username,
		Nickname:         userInfo.Nickname,
		Email:            userInfo.Email,
		IsAdmin:          userInfo.IsAdmin,
		TwoFactorEnabled: userInfo.TwoFactorEnabled,
		CreatedAt:        userInfo.CreatedAt,
	}, nil
}
