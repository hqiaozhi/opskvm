package totp

import (
	"context"

	v1 "opskvm/api/totp/v1"
	"opskvm/internal/logic/users"
)

func (c *ControllerV1) Enable(ctx context.Context, req *v1.EnableReq) (res *v1.EnableRes, err error) {
	userId := users.JwtInstance.GetUserIdFromCtx(ctx)
	if userId == 0 {
		return &v1.EnableRes{
			Success: false,
			Message: "unauthorized",
		}, nil
	}

	tempSecret, err := c.totp.GetTempSecret(ctx, userId)
	if err != nil {
		return &v1.EnableRes{
			Success: false,
			Message: "please call setup first to generate secret",
		}, nil
	}

	err = c.totp.Enable(ctx, userId, tempSecret, req.Code)
	if err != nil {
		return &v1.EnableRes{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	c.totp.ClearTempSecret(ctx, userId)

	return &v1.EnableRes{
		Success: true,
		Message: "TOTP enabled successfully",
	}, nil
}
