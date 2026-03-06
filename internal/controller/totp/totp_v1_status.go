package totp

import (
	"context"

	v1 "opskvm/api/totp/v1"
	"opskvm/internal/logic/users"
)

func (c *ControllerV1) Status(ctx context.Context, req *v1.StatusReq) (res *v1.StatusRes, err error) {
	userId := users.JwtInstance.GetUserIdFromCtx(ctx)
	if userId == 0 {
		return &v1.StatusRes{
			Enabled: false,
			Message: "unauthorized",
		}, nil
	}

	enabled, message, err := c.totp.Status(ctx, userId)
	if err != nil {
		return &v1.StatusRes{
			Enabled: false,
			Message: err.Error(),
		}, nil
	}

	return &v1.StatusRes{
		Enabled: enabled,
		Message: message,
	}, nil
}
