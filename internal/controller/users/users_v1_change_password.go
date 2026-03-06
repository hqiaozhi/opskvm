package users

import (
	"context"
	"strconv"
	"time"

	v1 "opskvm/api/users/v1"
	"opskvm/internal/logic/users"

	"github.com/gogf/gf/v2/net/ghttp"
)

func (c *ControllerV1) ChangePassword(ctx context.Context, req *v1.ChangePasswordReq) (res *v1.ChangePasswordRes, err error) {
	r := ghttp.RequestFromCtx(ctx)
	userIdStr := r.GetParam("userid")
	userId, err := strconv.Atoi(userIdStr.String())
	if err != nil {
		return nil, err
	}

	err = c.users.ChangePassword(ctx, userId, req.OldPassword, req.NewPassword)
	if err != nil {
		return nil, err
	}

	r2 := ghttp.RequestFromCtx(ctx)
	token := r2.GetCtxVar("token").String()
	if token != "" {
		expireTime, err := users.JwtInstance.GetTokenExpireTime(token)
		if err == nil {
			expireDuration := time.Until(expireTime)
			if expireDuration > 0 {
				users.JwtInstance.AddToBlacklist(token, expireDuration)
			}
		}
	}

	return &v1.ChangePasswordRes{}, nil
}
