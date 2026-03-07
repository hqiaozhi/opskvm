package totp

import (
	"context"

	v1 "opskvm/api/totp/v1"
	"opskvm/internal/logic/users"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

func (c *ControllerV1) Verify(ctx context.Context, req *v1.VerifyReq) (res *v1.VerifyRes, err error) {
	userId := users.JwtInstance.GetUserIdFromCtx(ctx)
	if userId == 0 {
		return &v1.VerifyRes{
			Success: false,
			Message: "unauthorized",
		}, nil
	}

	success, err := c.totp.Verify(ctx, userId, req.Code)
	if err != nil {
		tokenStr := g.RequestFromCtx(ctx).Header.Get("Authorization")
		if len(tokenStr) > 7 {
			tokenStr = tokenStr[7:]
			users.JwtInstance.AddToBlacklist(tokenStr, users.JwtInstance.ExpireHours)
		}
		return nil, gerror.New(err.Error())
	}

	if !success {
		tokenStr := g.RequestFromCtx(ctx).Header.Get("Authorization")
		if len(tokenStr) > 7 {
			tokenStr = tokenStr[7:]
			users.JwtInstance.AddToBlacklist(tokenStr, users.JwtInstance.ExpireHours)
		}
		return nil, gerror.New("invalid code")
	}

	return &v1.VerifyRes{
		Success: true,
		Message: "verification successful",
	}, nil
}
