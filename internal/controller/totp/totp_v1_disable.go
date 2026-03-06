package totp

import (
	"context"

	v1 "opskvm/api/totp/v1"
	"opskvm/internal/logic/users"
)

func (c *ControllerV1) Disable(ctx context.Context, req *v1.DisableReq) (res *v1.DisableRes, err error) {
	userId := users.JwtInstance.GetUserIdFromCtx(ctx)
	if userId == 0 {
		return &v1.DisableRes{
			Success: false,
			Message: "unauthorized",
		}, nil
	}

	success, err := c.totp.Disable(ctx, userId, req.Code, req.Password)
	if err != nil {
		return &v1.DisableRes{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	if success {
		return &v1.DisableRes{
			Success: true,
			Message: "TOTP disabled successfully",
		}, nil
	}

	return &v1.DisableRes{
		Success: false,
		Message: "failed to disable TOTP",
	}, nil
}
