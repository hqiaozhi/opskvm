package totp

import (
	"context"

	v1 "opskvm/api/totp/v1"
	"opskvm/internal/logic/users"

	"github.com/gogf/gf/v2/errors/gerror"
)

func (c *ControllerV1) Setup(ctx context.Context, req *v1.SetupReq) (res *v1.SetupRes, err error) {
	userId := users.JwtInstance.GetUserIdFromCtx(ctx)
	if userId == 0 {
		return nil, gerror.New("unauthorized")
	}

	secret, qrCode, err := c.totp.Setup(ctx, userId)
	if err != nil {
		return nil, err
	}

	return &v1.SetupRes{
		Secret: secret,
		QRCode: qrCode,
	}, nil
}
