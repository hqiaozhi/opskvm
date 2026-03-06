package users

import (
	"context"
	"time"

	v1 "opskvm/api/users/v1"
	"opskvm/internal/logic/users"

	"github.com/gogf/gf/v2/net/ghttp"
)

func (c *ControllerV1) Logout(ctx context.Context, req *v1.LogoutReq) (res *v1.LogoutRes, err error) {
	r := ghttp.RequestFromCtx(ctx)
	token := r.GetCtxVar("token").String()
	if token == "" {
		return nil, nil
	}

	expireTime, err := users.JwtInstance.GetTokenExpireTime(token)
	if err != nil {
		return nil, nil
	}

	expireDuration := time.Until(expireTime)
	if expireDuration > 0 {
		users.JwtInstance.AddToBlacklist(token, expireDuration)
	}

	return nil, nil
}
