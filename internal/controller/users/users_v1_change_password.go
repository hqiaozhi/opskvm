package users

import (
	"context"
	"time"

	v1 "opskvm/api/users/v1"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

func (c *ControllerV1) ChangePassword(ctx context.Context, req *v1.ChangePasswordReq) (res *v1.ChangePasswordRes, err error) {
	r := ghttp.RequestFromCtx(ctx)
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

	err = c.users.ChangePassword(ctx, targetUserId, req.OldPassword, req.NewPassword)
	if err != nil {
		return nil, err
	}

	if targetUserId == currentUserId {
		token := r.GetCtxVar("token").String()
		if token != "" {
			expireTime, err := c.users.JWT.GetTokenExpireTime(token)
			if err == nil {
				expireDuration := time.Until(expireTime)
				if expireDuration > 0 {
					c.users.JWT.AddToBlacklist(token, expireDuration)
				}
			}
		}
	}

	return &v1.ChangePasswordRes{}, nil
}
